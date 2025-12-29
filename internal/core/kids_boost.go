package core

type KidsBoostConfig struct {
	Enabled            bool
	SpeedMultiplier    float64 // 1.0-200.0, multiplies ball speed
	HeightMultiplier   float64 // 1.0-5.0, multiplies launch angle
	StraightnessBoost  float64 // 0.0-1.0, reduces HLA (1.0 = perfectly straight)
	PuttStraightness   float64 // 0.0-1.0, reduces putt HLA
}

type KidsBoostMiddleware struct {
	getConfig func() KidsBoostConfig
}

func NewKidsBoostMiddleware(configProvider func() KidsBoostConfig) *KidsBoostMiddleware {
	return &KidsBoostMiddleware{getConfig: configProvider}
}

func (m *KidsBoostMiddleware) TransformBallMetrics(metrics *BallMetrics) *BallMetrics {
	if metrics == nil {
		return metrics
	}

	config := m.getConfig()
	if !config.Enabled {
		return metrics
	}

	boosted := *metrics

	if metrics.ShotType == ShotTypePutt {
		boosted.HorizontalAngle = metrics.HorizontalAngle * (1.0 - config.PuttStraightness)
	} else {
		boosted.BallSpeedMPS = metrics.BallSpeedMPS * config.SpeedMultiplier
		boosted.VerticalAngle = metrics.VerticalAngle * config.HeightMultiplier
		boosted.HorizontalAngle = metrics.HorizontalAngle * (1.0 - config.StraightnessBoost)
	}

	return &boosted
}

func (m *KidsBoostMiddleware) TransformClubMetrics(metrics *ClubMetrics) *ClubMetrics {
	return metrics
}
