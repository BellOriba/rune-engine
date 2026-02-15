package handlers

import (
	"image/gif"
	"net/http"
	"strconv"
	"time"

	"github.com/BellOriba/rune-engine/internal/ascii"
	"github.com/BellOriba/rune-engine/internal/logger"
	"github.com/BellOriba/rune-engine/internal/worker"
	"github.com/gin-gonic/gin"
)

type ASCIIHandler struct {
	Pool *worker.Pool
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

	resultChan := make(chan string, 1)

	h.Pool.Submit(func() {
		conv := ascii.NewConverter(ascii.Options{
			TargetWidth: width,
			TargetHeight: height,
			Mode: mode,
		})
		resultChan <- conv.Convert(img)
	})

	select {
	case result := <- resultChan:
		c.String(http.StatusOK, result)
	case <-time.After(10 * time.Second):
		log.Warn("timeout no processamento da imagem")
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "servidor ocupado, tente novamente"})
	
	}
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

	conv := ascii.NewConverter(ascii.Options{
		TargetWidth: width,
		TargetHeight: height,
		Mode: mode,
	})

	log.Info("iniciando streaming de GIF", "frames", len(g.Image), "width", width)

	for i := range g.Image {
		h.Pool.Submit(func () {
			asciiFrame := conv.Convert(g.Image[i])
			frameChans[i] <- asciiFrame
		})
	}

	for i := range totalFrames {	
		asciiFrame := <-frameChans[i]
		c.Writer.Write([]byte("data: \033[H"))
		c.Writer.Write([]byte(asciiFrame))
		c.Writer.Write([]byte("\n\n"))
		c.Writer.Flush()

		time.Sleep(time.Duration(g.Delay[i]) * 10 * time.Millisecond)
	}
	log.Info("streaming de GIF finalizado com sucesso")
}

