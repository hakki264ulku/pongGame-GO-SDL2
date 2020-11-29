package main

// paddles
// ball
// keyboard states to control paddle 1, and ai updater for the paddle2 which could be never beaten
// pos as a general

import (
	"fmt"
	"image/png"
	"os"
	"strconv"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Go's built il function for sorting needs three things: 1-swap(i,j), 2-Len(Arr), 3-Less(a,b)

const winWidth int = 800
const winHeight int = 600

type rgba struct {
	r, g, b byte
}

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type Score struct {
	val      int
	renderer *sdl.Renderer
	font     *ttf.Font
	rect     *sdl.Rect
}

func (score *Score) update() {
	textSurface, err := score.font.RenderUTF8Solid(strconv.Itoa(score.val), sdl.Color{R: 255, G: 255, B: 255})
	if err != nil {
		panic("Error: there is a problem with score updating!")
	}
	textTexture, err := score.renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		panic("Error: there is a problem with score updating!")
	}
	score.renderer.Copy(textTexture, nil, score.rect)
}

func textToTexture(text string, font *ttf.Font, c *rgba, renderer *sdl.Renderer) *sdl.Texture {
	textSurface, err := font.RenderUTF8Solid(text, sdl.Color{R: c.r, G: c.g, B: c.b})
	if err != nil {
		panic("error in textTotexture")
	}

	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		panic("error in textTotexture")
	}
	renderer.Copy(textTexture, nil, nil)
	return textTexture
}

type Pos struct {
	x, y int
}

type Paddle struct {
	Pos
	speed int
	tex   *sdl.Texture
	w, h  int
}

func (paddle *Paddle) draw(renderer *sdl.Renderer) {
	renderer.Copy(paddle.tex, nil, &sdl.Rect{int32(paddle.x), int32(paddle.y), int32(paddle.w), int32(paddle.h)})
}

func (paddle *Paddle) update(keyState []uint8, elapsedTime float32) {
	if keyState[sdl.SCANCODE_UP] != 0 && paddle.y > 0 && paddle.y+paddle.h < winHeight {
		paddle.y -= int(float32(paddle.speed) * elapsedTime / 15)
		if paddle.y < 0 {
			paddle.y = 1
		}
	}

	if keyState[sdl.SCANCODE_DOWN] != 0 && paddle.y >= 0 && paddle.y < winHeight {
		paddle.y += int(float32(paddle.speed) * elapsedTime / 15)
		if paddle.y+paddle.h > winHeight {
			paddle.y = winHeight - paddle.h - 1
		}
	}
}

func (paddle *Paddle) aiUpdate(ball *Ball) {
	paddle.y = ball.y - paddle.h/2
	if paddle.y < 0 {
		paddle.y = 1
	}
	if paddle.y+paddle.h > winHeight {
		paddle.y = winHeight - paddle.h - 1
	}
}

type Ball struct {
	Pos
	xv, yv float32
	tex    *sdl.Texture
	w, h   int
}

func (ball *Ball) draw(renderer *sdl.Renderer) {
	renderer.Copy(ball.tex, nil, &sdl.Rect{int32(ball.x), int32(ball.y), 32, 32})
}

func (ball *Ball) update(paddleLeft, paddleRight Paddle, elapsedTime float32, score1, score2 *Score, audioState *audioState) {
	ball.x += int(float32(ball.xv) * elapsedTime / 10)
	ball.y += int(float32(ball.yv) * elapsedTime / 10)

	if ball.y < 0 {
		ball.yv = -ball.yv
		ball.y += int(ball.yv)
	}
	if ball.y+ball.h >= winHeight-1 {
		ball.yv = -ball.yv
		ball.y += int(ball.yv)
	}

	if ball.y > paddleLeft.y && ball.y < paddleLeft.y+paddleLeft.h && ball.x <= paddleLeft.x+paddleLeft.w {
		ball.yv = -ball.yv
		ball.xv = -ball.xv
		ball.x += int(ball.xv)
		sdl.ClearQueuedAudio(audioState.deviceID)

		// when a balloon hits you just queueu the audio data and stop pausing the audio device to run that data
		sdl.QueueAudio(audioState.deviceID, audioState.explosionBytes)
		sdl.PauseAudioDevice(audioState.deviceID, false)
	}

	if ball.y > paddleRight.y && ball.y < paddleRight.y+paddleRight.h && ball.x+ball.w >= paddleRight.x {
		ball.yv = -ball.yv
		ball.xv = -ball.xv
		ball.x += int(ball.xv)
		sdl.ClearQueuedAudio(audioState.deviceID)

		// when a balloon hits you just queueu the audio data and stop pausing the audio device to run that data
		sdl.QueueAudio(audioState.deviceID, audioState.explosionBytes)
		sdl.PauseAudioDevice(audioState.deviceID, false)
	}

	// left ball loses
	if ball.x <= -1 {
		ball.Pos = Pos{400, 300}
		ball.yv = -ball.yv
		ball.xv = -ball.xv
		score2.val++
	}

	// // right ball loses but it should never lose
	if ball.x+ball.w >= winWidth+1 {
		ball.Pos = Pos{400, 300}
		ball.yv = -ball.yv
		ball.xv = -ball.xv
	}

}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func imgFileToTexture(renderer *sdl.Renderer, fileName string) *sdl.Texture {

	infile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}

	tex := pixelsToTexture(renderer, pixels, w, h)

	// for alpha blending with sdl
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}

	return tex

}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow("Testing SDL2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err, " occured")
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err, " occured")
		return
	}
	defer renderer.Destroy()

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err, " occured")
		return
	}
	defer tex.Destroy()

	if ttf.Init() != nil {
		fmt.Println(err, " occured")
		return
	}
	defer ttf.Quit()

	font, err := ttf.OpenFont("OpenSans-Regular.ttf", 128)
	if err != nil {
		fmt.Println("Error: Problem with opening the font")
	}

	explosionBytes, audioSpec := sdl.LoadWAV("pat.wav")
	audioID, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
	if err != nil {
		panic(err)
	}
	defer sdl.FreeWAV(explosionBytes)

	audioState := audioState{explosionBytes, audioID, audioSpec}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var elapsedTime float32
	keyState := sdl.GetKeyboardState()

	paddle := Paddle{Pos{10, 10}, 10, imgFileToTexture(renderer, "fancy-paddle-green.png"), 32, 128}
	paddle2 := Paddle{Pos{winWidth - 10 - 32, 300}, 10, imgFileToTexture(renderer, "fancy-paddle-blue.png"), 32, 128}
	ball := Ball{Pos{400, 300}, 5, 7, imgFileToTexture(renderer, "fancy-ball.png"), 32, 32}
	score1 := Score{0, renderer, font, &sdl.Rect{340, 10, 40, 110}}
	score2 := Score{0, renderer, font, &sdl.Rect{410, 10, 40, 110}}

	gameOverTexture := textToTexture("GAME OVER", font, &rgba{255, 255, 255}, renderer)
	playAgainTexture := textToTexture("Press Space to play again", font, &rgba{100, 200, 220}, renderer)

	court := imgFileToTexture(renderer, "fancy-court.png")
	renderer.Copy(court, nil, nil)

	// Game Loop //
	for {

		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) { // theEvent := event.(type) //remember this
			case *sdl.QuitEvent:
				return
			}
		}

		if score2.val > 3 {
			for {

				for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
					switch event.(type) { // theEvent := event.(type) //remember this
					case *sdl.QuitEvent:
						return
					}
				}

				if keyState[sdl.SCANCODE_SPACE] != 0 {
					score1.val = 0
					score2.val = 0
					ball.y = 400
					ball.x = 300
					break
				}

				renderer.Copy(court, nil, nil)
				// renderer.Copy(endPage, nil, nil)
				renderer.Copy(gameOverTexture, nil, &sdl.Rect{80, 0, 80 * 8, 70 * 8})
				renderer.Copy(playAgainTexture, nil, &sdl.Rect{300, 500, 200, 50})

				renderer.Present()
			}
		}

		paddle.update(keyState, elapsedTime)
		paddle2.aiUpdate(&ball)
		ball.update(paddle, paddle2, elapsedTime, &score1, &score2, &audioState)

		renderer.Copy(court, nil, nil)
		paddle.draw(renderer)
		paddle2.draw(renderer)
		ball.draw(renderer)
		score1.update()
		score2.update()

		renderer.Present()

		// related to fps perfection
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//fmt.Println("ms per frame: ", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
