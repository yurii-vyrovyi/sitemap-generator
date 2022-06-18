package reporter

type (
	Reporter struct {
		config Config
	}

	Config struct {
		FileName string
	}
)

func New(config Config) *Reporter {
	return &Reporter{
		config: config,
	}
}

func (r *Reporter) Save() error {
	return nil
}
