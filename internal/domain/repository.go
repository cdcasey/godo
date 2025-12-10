package domain

type UserRepository interface {
	Create(user *User) error
	GetByEmail(email *string) (*User, error)
	GetByID(id *string) (*User, error)
}

type TodoRepository interface {
	Create(todo *Todo) error
	GetByID(id *string) (*Todo, error)
	GetByUserID(id *string) ([]*Todo, error)
	GetAll() ([]*Todo, error)
	Update(id *string) error
	Delete(id *string) error
}
