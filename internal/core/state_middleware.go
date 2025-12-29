package core

type StateMiddleware interface {
	TransformBallMetrics(*BallMetrics) *BallMetrics
	TransformClubMetrics(*ClubMetrics) *ClubMetrics
}

type CompositeMiddleware struct {
	middlewares []StateMiddleware
}

func NewCompositeMiddleware() *CompositeMiddleware {
	return &CompositeMiddleware{
		middlewares: make([]StateMiddleware, 0),
	}
}

func (c *CompositeMiddleware) Add(m StateMiddleware) {
	c.middlewares = append(c.middlewares, m)
}

func (c *CompositeMiddleware) TransformBallMetrics(metrics *BallMetrics) *BallMetrics {
	result := metrics
	for _, m := range c.middlewares {
		result = m.TransformBallMetrics(result)
	}
	return result
}

func (c *CompositeMiddleware) TransformClubMetrics(metrics *ClubMetrics) *ClubMetrics {
	result := metrics
	for _, m := range c.middlewares {
		result = m.TransformClubMetrics(result)
	}
	return result
}
