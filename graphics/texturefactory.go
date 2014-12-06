package graphics

import (
	"image"
)

type Filter int

const (
	FilterNearest Filter = iota
	FilterLinear
)

type TextureId int

// A render target is essentially same as a texture, but it is assumed that the
// all alpha of a render target is maximum.
type RenderTargetId int

var currentTextureFactory TextureFactory

type TextureFactory interface {
	CreateRenderTarget(width, height int, filter Filter) (RenderTargetId, error)
	CreateTexture(img image.Image, filter Filter) (TextureId, error)
}

func SetTextureFactory(textureFactory TextureFactory) {
	currentTextureFactory = textureFactory
}

func CreateRenderTarget(width, height int, filter Filter) (RenderTargetId, error) {
	if currentTextureFactory == nil {
		panic("graphics.CreateRenderTarget: currentTextureFactory is not set.")
	}
	return currentTextureFactory.CreateRenderTarget(width, height, filter)
}

func CreateTexture(img image.Image, filter Filter) (TextureId, error) {
	if currentTextureFactory == nil {
		panic("graphics.CreateTexture: currentTextureFactory is not set")
	}
	return currentTextureFactory.CreateTexture(img, filter)
}