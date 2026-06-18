package Engine

const (
	Interpolation_Type_Linear = iota
)

const (
	// WARNING!!!!!! MUST BE SAME AS TIMER STATE!!!
	InterpolationState_Running = iota
	InterpolationState_Paused
	InterpolationState_Finished
)

type Interpolation[T any] struct {
	InterpolationType int
	Loop              bool

	inverseInterpolationParameter bool

	InterpolateStartValue T
	InterpolateValue      *T
	InterpolateToValue    *T

	InterpolationCalculator func(t *float64, startValue T, currentValue *T, futureValue *T)

	interpolationTimerPoolItem *PoolItem[Timer]
}

func (interpolation *Interpolation[T]) GetInterpolationParameter() float64 {
	t := float64(float64(interpolation.interpolationTimerPoolItem.Item.TimePassedSinceStart()) / float64(interpolation.interpolationTimerPoolItem.Item.TimerTotalDuration))
	if interpolation.inverseInterpolationParameter {
		return 1.0 - t
	}
	return float64(float64(interpolation.interpolationTimerPoolItem.Item.TimePassedSinceStart()) / float64(interpolation.interpolationTimerPoolItem.Item.TimerTotalDuration))
}

func (interpolation *Interpolation[T]) GetInterpolationState() int {
	return interpolation.interpolationTimerPoolItem.Item.TimerState
}

type InterpolationSystem[T any] struct {
	InterpolationsPool Pool[Interpolation[T]]
	TimerSystem        TimerSystem
}

func (interpolationSystem *InterpolationSystem[T]) InitWithInterpolationsInPool(interpolationSystemName string, totalNumInterpolationsToCreateInPool int) {

	interpolationSystem.InterpolationsPool.InitPool(interpolationSystemName, totalNumInterpolationsToCreateInPool)
	interpolationSystem.TimerSystem.InitWithTimers(interpolationSystemName+"'s Timer System", totalNumInterpolationsToCreateInPool)
}

func (interpolationSystem *InterpolationSystem[T]) CreateNewInterpolation(interpolateStartValue T, interpolateValue *T, interpolateToValue *T, interpolateInTime float64, shouldLoop bool, pingPong bool, InterpolationCalculator func(t *float64, startValue T, currentValue *T, futureValue *T), OnFinishInterpolation func()) *PoolItem[Interpolation[T]] {

	curInterpolationPoolItem := interpolationSystem.InterpolationsPool.GetAnUnusedItemFromPool()

	curInterpolationPoolItem.Item.InterpolateStartValue = interpolateStartValue
	curInterpolationPoolItem.Item.InterpolateValue = interpolateValue
	curInterpolationPoolItem.Item.InterpolateToValue = interpolateToValue

	curInterpolationPoolItem.Item.inverseInterpolationParameter = false

	curInterpolationPoolItem.Item.InterpolationCalculator = InterpolationCalculator

	curInterpolationPoolItem.Item.interpolationTimerPoolItem = interpolationSystem.TimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(interpolateInTime, shouldLoop || pingPong, func() {
		if pingPong {
			curInterpolationPoolItem.Item.inverseInterpolationParameter = !curInterpolationPoolItem.Item.inverseInterpolationParameter
		}
		OnFinishInterpolation()
	})

	return curInterpolationPoolItem
}

func (interpolationSystem *InterpolationSystem[T]) UpdateAllInterpolationDeltasAndStates(dt float64) {

	interpolationSystem.TimerSystem.UpdateAllTimerDeltasAndStates(dt)

	interpolationSystem.InterpolationsPool.PerformOperationOnAlivePoolItems(func(curInterpolationPoolItem *PoolItem[Interpolation[T]]) {
		curInterpolationParameter := curInterpolationPoolItem.Item.GetInterpolationParameter()
		curInterpolationPoolItem.Item.InterpolationCalculator(&curInterpolationParameter, curInterpolationPoolItem.Item.InterpolateStartValue, curInterpolationPoolItem.Item.InterpolateValue, curInterpolationPoolItem.Item.InterpolateToValue)
	})

	interpolationSystem.InterpolationsPool.PerformOperationOnAlivePoolItemsBackwards(func(curInterpolationPoolItem *PoolItem[Interpolation[T]]) {
		if curInterpolationPoolItem.Item.GetInterpolationState() == InterpolationState_Finished {
			*curInterpolationPoolItem.Item.InterpolateValue = *curInterpolationPoolItem.Item.InterpolateToValue
			interpolationSystem.KillInterpolation(curInterpolationPoolItem)
		}
	})
}

func (interpolationSystem *InterpolationSystem[T]) KillInterpolation(interpolationPoolItem *PoolItem[Interpolation[T]]) {
	interpolationSystem.InterpolationsPool.KillItemInPool(interpolationPoolItem)
}
