package content

type WorldBookRepository struct {
	baseDirectory string
}

func NewWorldBookRepository(baseDirectory string) *WorldBookRepository {
	return &WorldBookRepository{baseDirectory}
}
