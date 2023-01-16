package crdt

type (
	discussion struct {
		name     string
		messages []message
	}

	Discussion interface {
		AddMessage()
		UpdateList()
	}
)
