package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestService(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

type ServiceSuite struct {
	suite.Suite
}

func (s *ServiceSuite) TestAnalyze_empty_input() {
	srv := NewService()

	out, err := srv.Analyze(1, time.Now(), []InputItem{}...)

	s.NoError(err)
	s.Equal(Output{}, out)
}
