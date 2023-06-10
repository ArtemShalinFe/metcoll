package sleepstepper

import "time"

type SleepSteper struct {
	sleep    int
	step     int
	maxSleep int
}

func NewSleepStepper(sleep int, step int, maxSleep int) *SleepSteper {

	if sleep == 0 {
		sleep = 1
	}

	if step == 0 {
		step = 2
	}

	if maxSleep == 0 {
		maxSleep = 5
	}

	return &SleepSteper{
		sleep:    sleep,
		step:     step,
		maxSleep: maxSleep,
	}

}

func (ss *SleepSteper) doStep() {
	ss.sleep += ss.step
}

func (ss *SleepSteper) canSleep() bool {
	return ss.sleep < ss.maxSleep
}

func (ss *SleepSteper) Sleep() bool {

	if !ss.canSleep() {
		return false
	}

	time.Sleep(time.Duration(ss.sleep) * time.Second)
	ss.doStep()

	return true

}
