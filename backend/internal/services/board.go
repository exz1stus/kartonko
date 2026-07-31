package services

type BoardService interface {
}

type boardService struct {
}

func NewBoardService() BoardService {
	return &boardService{}
}
