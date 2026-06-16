package Engine

import (
	"time"
)

const (
	// WARNING!!!!!! MUST BE SAME AS INTERPOLATION STATE!!!
	TimerState_Running = iota
	TimerState_Paused
	TimerState_Finished
)

type Timer struct {
	TimerTotalDuration     time.Duration
	TimerDurationRemaining time.Duration
	Loop                   bool
	TimerState             int
	OnFinish               func()
}

func (timer *Timer) TimePassedSinceStart() time.Duration {
	return timer.TimerTotalDuration - timer.TimerDurationRemaining
}

func (timer *Timer) SetDuration(duration time.Duration) {
	timer.TimerTotalDuration = duration
	timer.TimerDurationRemaining = duration
}

func (timer *Timer) PauseTimer() {
	timer.TimerState = TimerState_Paused
}

func (timer *Timer) UnPauseTimer() {
	timer.TimerState = TimerState_Running
}

func (timer *Timer) RestartTimer() {
	timer.TimerState = TimerState_Running
	timer.TimerDurationRemaining = timer.TimerTotalDuration
}

func (timer *Timer) ForceEndCurrentLoopOfTimerForNextUpdate() {
	timer.TimerDurationRemaining = 0.0
}

type TimerSystem struct {
	timerPool      Pool[Timer]
	lastUpdateTime time.Time
	delta          time.Duration
}

func (timerSystem *TimerSystem) InitWithTimers(timerSystemName string, totalNumTimersToCreate int) {
	timerSystem.timerPool.InitPool(timerSystemName, totalNumTimersToCreate)
	timerSystem.lastUpdateTime = time.Now()
	timerSystem.delta = 0.0 * time.Second
}

func (timerSystem *TimerSystem) SetTimerFromPoolWithDurationLoopAndFunc(duration time.Duration, shouldLoop bool, timerOnFinish func()) *PoolItem[Timer] {
	poolItemToReturn := timerSystem.timerPool.GetAnUnusedItemFromPool()

	poolItemToReturn.Item.SetDuration(duration)
	poolItemToReturn.Item.Loop = shouldLoop
	poolItemToReturn.Item.OnFinish = timerOnFinish
	poolItemToReturn.Item.TimerState = TimerState_Running

	return poolItemToReturn
}

func (timerSystem *TimerSystem) UpdateAllTimerDeltasAndStates() {

	timerSystem.delta = time.Since(timerSystem.lastUpdateTime)
	timerSystem.lastUpdateTime = time.Now()

	timerSystem.timerPool.PerformOperationOnAlivePoolItems(func(curTimer *PoolItem[Timer]) {

		if curTimer.Item.TimerState != TimerState_Paused {

			curTimer.Item.TimerDurationRemaining -= timerSystem.delta

			if curTimer.Item.TimerDurationRemaining <= 0.0 {

				curTimer.Item.OnFinish()

				if curTimer.Item.Loop {
					curTimer.Item.SetDuration(curTimer.Item.TimerTotalDuration)
					curTimer.Item.TimerState = TimerState_Running
				} else {
					curTimer.Item.TimerState = TimerState_Finished
				}
			}
		}
	})

	timerSystem.timerPool.PerformOperationOnAlivePoolItemsBackwards(func(curTimer *PoolItem[Timer]) {
		if curTimer.Item.TimerState == TimerState_Finished {
			timerSystem.KillTimer(curTimer)
		}
	})
}

func (timerSystem *TimerSystem) KillTimer(timerPoolItem *PoolItem[Timer]) {
	timerSystem.timerPool.KillItemInPool(timerPoolItem)
}
