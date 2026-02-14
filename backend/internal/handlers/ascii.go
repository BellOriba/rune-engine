package handlers

import (
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
		log.Error("falha ao descodificar imagem", "error", err)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "formato de imagem não suportado"})
		return
	}

	widthStr := c.DefaultQuery("width", "100")
	width, _ := strconv.Atoi(widthStr)
	mode := c.DefaultQuery("mode", "plain")

	resultChan := make(chan string, 1)

	h.Pool.Submit(func() {
		conv := ascii.NewConverter(ascii.Options{
			TargetWidth: width,
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

