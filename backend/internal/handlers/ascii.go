package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/gif"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/BellOriba/rune-engine/internal/ascii"
	"github.com/BellOriba/rune-engine/internal/cache"
	"github.com/BellOriba/rune-engine/internal/logger"
	"github.com/BellOriba/rune-engine/internal/worker"
	"github.com/gin-gonic/gin"
)

type ASCIIHandler struct {
	Pool *worker.Pool
	Cache *cache.Cache
}

func generateCacheKey(fileContent []byte, width, height int, mode string) string {
	hash := sha256.Sum256(fileContent)
	return "img:" + hex.EncodeToString(hash[:]) + ":" + strconv.Itoa(width) + ":" + strconv.Itoa(height) + ":" + mode
}

func (h *ASCIIHandler) ConvertImage(c *gin.Context) {
	log := logger.FromContext(c)

	file, err := c.FormFile("image")
	if err != nil {
		log.Error("falha ao receber imagem", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "imagem é obrigatória"})
		return
	}

	src, err := file.Open()
	if err != nil {
		log.Error("falha ao abrir imagem", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao processar imagem"})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		log.Error("falha ao ler bytes da imagem", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	if _, err := src.Seek(0, 0); err != nil {
		log.Error("falha ao resetar ponteiro do arquivo", "error", err)
	}

	img, _, err := ascii.Decode(src)
	if err != nil {
		log.Error("falha ao decodificar imagem", "error", err)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "formato de imagem não suportado"})
		return
	}

	widthStr := c.DefaultQuery("width", "180")
	width, _ := strconv.Atoi(widthStr)
	heightStr := c.DefaultQuery("height", "0")
	height, _ := strconv.Atoi(heightStr)
	mode := c.DefaultQuery("mode", "plain")

	cacheKey := generateCacheKey(fileBytes, width, height, mode)

	if cached, err := h.Cache.Get(c.Request.Context(), cacheKey); err == nil {
		c.Header("X-Cache", "HIT")
		c.String(http.StatusOK, cached)
		return
	}

	resultChan := make(chan string, 1)

	ctx := c.Request.Context()

	h.Pool.Submit(func() {
		select {
		case <-ctx.Done():
			return
		default:
			conv := ascii.NewConverter(ascii.Options{
				TargetWidth: width,
				TargetHeight: height,
				Mode: mode,
			})
			result := conv.Convert(img)
			h.Cache.Set(context.Background(), cacheKey, result, 24*time.Hour)

			select {
			case resultChan <- result:
			case <-ctx.Done():
			}
		}
	})

	select {
	case result := <- resultChan:
		c.String(http.StatusOK, result)
	case <-ctx.Done():
		log.Warn("cliente desconectou, cancelando resposta")
	case <-time.After(10 * time.Second):
		log.Warn("timeout no processamento da imagem")
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "servidor ocupado, tente novamente"})	
	}
}

func (h *ASCIIHandler) processFrameWithCache(ctx context.Context, frame image.Image, opts ascii.Options, log *slog.Logger) string {
	var pix []byte
	switch img := frame.(type) {
	case *image.RGBA:
		pix = img.Pix
	case *image.Paletted:
		pix = img.Pix
	default:
		return ascii.NewConverter(opts).Convert(frame)
	}

	hash := sha256.Sum256(pix)
	cacheKey := hex.EncodeToString(hash[:]) + ":" + strconv.Itoa(opts.TargetWidth) + ":" + strconv.Itoa(opts.TargetHeight) + ":" + opts.Mode

	if h.Cache != nil {
		if cached, err := h.Cache.Get(ctx, cacheKey); err == nil {
			return cached
		}
	}

	conv := ascii.NewConverter(opts)
	result := conv.Convert(frame)

	if h.Cache != nil {
		go func() {
			err := h.Cache.Set(context.Background(), cacheKey, result, 24*time.Hour)
			if err != nil {
				log.Debug("falha ao salvar frame no cache", "error", err)
			}
		}()
	}

	return result
}

func (h *ASCIIHandler) StreamGIF(c *gin.Context) {
	log := logger.FromContext(c)

	file, err := c.FormFile("image")
	if err != nil {
		log.Error("falha ao receber GIF", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "imagem é obrigatória"})
		return
	}

	f, _ := file.Open()
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		log.Error("falha ao decodificar GIF", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "falha ao decodificar GIF"})
		return
	}

	widthStr := c.DefaultQuery("width", "180")
	width, _ := strconv.Atoi(widthStr)
	heightStr := c.DefaultQuery("height", "0")
	height, _ := strconv.Atoi(heightStr)
	mode := c.DefaultQuery("mode", "plain")

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	totalFrames := len(g.Image)
	frameChans := make([]chan string, totalFrames)
	for i := range frameChans {
		frameChans[i] = make(chan string, 1)
	}

	opts := ascii.Options{
		TargetWidth: width,
		TargetHeight: height,
		Mode: mode,
	}

	log.Info("iniciando streaming de GIF", "frames", len(g.Image), "width", width)

	for i := range g.Image {
		frameIndex := i
		frameImg := g.Image[frameIndex]

		h.Pool.Submit(func() {
			asciiFrame := h.processFrameWithCache(c.Request.Context(), frameImg, opts, log)
			frameChans[frameIndex] <- asciiFrame
		})
	}

	for i := range totalFrames {
		select {
		case <-c.Request.Context().Done():
			log.Warn("cliente desconectou, interrompendo streaming de GIF")
			return
		case asciiFrame := <-frameChans[i]:
			c.Writer.Write([]byte("data: \033[H"))
			c.Writer.Write([]byte(asciiFrame))
			c.Writer.Write([]byte("\n\n"))
			c.Writer.Flush()

			time.Sleep(time.Duration(g.Delay[i]) * 10 * time.Millisecond)
		}
	}
	log.Info("streaming de GIF finalizado com sucesso")
}

