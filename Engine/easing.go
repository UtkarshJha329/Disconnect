package Engine

func LinearInterpolationInt(interpolationParameter *float64, startValue int, curValueToChange *int, finalValue *int) {
	newValue := (startValue) + int(float64(*finalValue-startValue)*(*interpolationParameter))
	*curValueToChange = newValue
}

func LinearInterpolationFloat(interpolationParameter *float64, startValue float64, curValueToChange *float64, finalValue *float64) {
	newValue := (startValue) + (*finalValue-startValue)*(*interpolationParameter)
	*curValueToChange = newValue
}

func LinearInterpolationVector2(interpolationParameter *float64, startValue Vector2, curValueToChange *Vector2, finalValue *Vector2) {
	totalDeltaFromStartToFinish := finalValue.Subtract(&startValue)
	scaledDelta := totalDeltaFromStartToFinish.MultiplyFloat(*interpolationParameter)
	newValue := startValue.Add(&scaledDelta)
	*curValueToChange = newValue
}
