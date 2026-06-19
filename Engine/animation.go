package Engine

import "github.com/hajimehoshi/ebiten/v2"

type Animation struct {
	SpriteSheet *ebiten.Image

	AnimationIndex int

	NumAnimationFrames int
	FrameSize          Vector2

	CurrentAnimationFrame int

	ShouldLoop            bool
	TotalAnimationDuraion float64
	AnimationTimerSystem  TimerSystem
}

type AnimatedSprite struct {
	CurAnimationIndex  int
	TotalNumAnimations int
	Animations         map[int]*Animation
}

func (animatedSprite *AnimatedSprite) CreateNewAnimation(
	spriteSheetSrc string,
	numAnimationFrames int,
	frameSize Vector2,
	shouldLoop bool,
	duration float64,
) int {
	newAnimation := Animation{
		SpriteSheet:           GetImageFromFile(spriteSheetSrc),
		AnimationIndex:        animatedSprite.TotalNumAnimations,
		NumAnimationFrames:    numAnimationFrames,
		FrameSize:             frameSize,
		CurrentAnimationFrame: 0,
		ShouldLoop:            shouldLoop,
		TotalAnimationDuraion: duration,
	}
	newAnimation.AnimationTimerSystem.InitWithTimers(spriteSheetSrc+"'s Animation Timer System.", 1)
	newAnimation.AnimationTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
		newAnimation.TotalAnimationDuraion/float64(newAnimation.NumAnimationFrames),
		newAnimation.ShouldLoop,
		func() {
			newAnimation.CurrentAnimationFrame++
			newAnimation.CurrentAnimationFrame %= newAnimation.NumAnimationFrames
		},
	)

	animatedSprite.Animations[newAnimation.AnimationIndex] = &newAnimation

	animatedSprite.TotalNumAnimations++

	return newAnimation.AnimationIndex
}

func (animatedSprite *AnimatedSprite) ChangeCurrentAnimationToAnimationIndex(newAnimationIndex int) {

	if animatedSprite.CurAnimationIndex != newAnimationIndex {
		if curAnimation, ok := animatedSprite.Animations[animatedSprite.CurAnimationIndex]; ok {
			curAnimation.AnimationTimerSystem.ForceEndAllTimersForNextUpdate()
		}

		animatedSprite.CurAnimationIndex = newAnimationIndex
	}
}

func AnimationEntityUpdateFunc(world *World, dt float64) {
	for e, curAnimatedSprite := range world.AnimatedSprites {

		curAnimation := curAnimatedSprite.Animations[curAnimatedSprite.CurAnimationIndex]
		curAnimation.AnimationTimerSystem.UpdateAllTimerDeltasAndStates(dt)

		curSprite := world.Sprites[e]
		curSprite.Tex = curAnimation.SpriteSheet
		// curSprite.FramePosition = (*curAnimatedSprite.FramePosition)[curAnimatedSprite.CurrentAnimationFrame]
		curSprite.FramePosition = Vector2{
			X: curAnimation.FrameSize.X * float64(curAnimation.CurrentAnimationFrame),
			Y: 0 * float64(curAnimation.CurrentAnimationFrame),
		}

		curSprite.FrameSize = curAnimation.FrameSize

	}
}
