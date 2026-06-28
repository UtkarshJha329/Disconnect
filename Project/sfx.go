package Project

import (
	"Disconnect/Engine"
	"io"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

var AudioContext *audio.Context
var SoundEffects map[string]*[]byte

const SampleRate = 48000

func loadSound(path string) ([]byte, error) {
	file, err := Assets.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decodedStream, err := wav.DecodeWithoutResampling(file)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(decodedStream)
}

func AddSoundEffectToMap(soundEffectPath string) {
	sfxBytes, err := loadSound(soundEffectPath)
	if err != nil {
		log.Fatalf("failed to load sound from folder: %v", err)
	}

	SoundEffects[soundEffectPath] = &sfxBytes
}

func SFXInitFunc(world *Engine.World) {

	SoundEffects = make(map[string]*[]byte)

	AddSoundEffectToMap("Assets/Sfx/Checkpoint.wav")
	AddSoundEffectToMap("Assets/Sfx/Hit.wav")
	AddSoundEffectToMap("Assets/Sfx/Jump.wav")
	AddSoundEffectToMap("Assets/Sfx/Menuselect.wav")
	AddSoundEffectToMap("Assets/Sfx/PowerUp.wav")
	AddSoundEffectToMap("Assets/Sfx/Dash.wav")

}

func PlaySoundEffect(sfxPath string) {

	if SoundEffects[sfxPath] == nil {
		return
	}

	soundEffectPlayer := AudioContext.NewPlayerFromBytes(*SoundEffects[sfxPath])
	soundEffectPlayer.Play()
}

func SFXUpdateFunc(world *Engine.World, dt float64) {}
