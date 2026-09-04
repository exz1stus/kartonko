package tag

type TagPostBatchRequest struct {
	Names []string
}

type TagPostRequest struct {
	Name string
}

type TagResponse struct {
	ID   uint
	Name string
}

type TagBatchResponse struct {
	Successes []TagResponse
	Failures  []struct {
		Name  string
		Error string
	}
}
