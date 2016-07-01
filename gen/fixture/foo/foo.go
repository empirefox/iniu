package foo

import "github.com/jinzhu/gorm"

//type Model struct {
//	ID        uint `gorm:"primary_key"`
//	CreatedAt time.Time
//	UpdatedAt time.Time
//	DeletedAt *time.Time `sql:"index"`
//}

type Foo struct {
	Id  uint   `SA:"#"`
	Bar string `VIEW:";lmax(16)" SA:"+"`
}

type Alice struct {
	ID   uint   `SA:"#"`
	Name string `MGR:";lmax(16)" SA:"+"`
}

// bob, doc
type Bob struct {
	*Foo
	gorm.Model
	Name  string `MGR:";lmax(16)" SA:"+"`
	Alice Alice
}

type Boys []Bob

type Boyss []*Bob

type Int int

type Ints []int

type IntMap map[int]interface{}

type BobMap map[string]Bob

type BobsMap map[int]*Bob
