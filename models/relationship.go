package models

import (
	"github.com/lifei6671/godoc/conf"
	"github.com/astaxie/beego/orm"
	"errors"
	"github.com/astaxie/beego/logs"
)

type Relationship struct {
	RelationshipId int	`orm:"pk;auto;unique;column(relationship_id)" json:"relationship_id"`
	MemberId int		`orm:"column(member_id);type(int)" json:"member_id"`
	BookId int		`orm:"column(book_id);type(int)" json:"book_id"`
	// RoleId 角色：0 创始人(创始人不能被移除) / 1 管理员/2 编辑者/3 观察者
	RoleId int		`orm:"column(role_id);type(int)" json:"role_id"`
}


// TableName 获取对应数据库表名.
func (m *Relationship) TableName() string {
	return "relationship"
}
func (m *Relationship) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

// TableEngine 获取数据使用的引擎.
func (m *Relationship) TableEngine() string {
	return "INNODB"
}

// 联合唯一键
func (u *Relationship) TableUnique() [][]string {
	return [][]string{
		[]string{"MemberId", "BookId"},
	}
}

func NewRelationship() *Relationship {
	return &Relationship{}
}

func (m *Relationship) Find(id int) (*Relationship,error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("relationship_id",id).One(m)
	return m,err
}

func (m *Relationship) UpdateRoleId(book_id,member_id, role_id int) (*Relationship,error) {
	o := orm.NewOrm()
	book := NewBook()
	book.BookId = book_id

	if err := o.Read(book); err != nil {
		logs.Error("UpdateRoleId => ", err)
		return m,errors.New("项目不存在")
	}
	err := o.QueryTable(m.TableNameWithPrefix()).Filter("member_id",member_id).Filter("book_id",book_id).One(m)

	if err == orm.ErrNoRows {
		m = NewRelationship()
		m.BookId = book_id
		m.MemberId = member_id
		m.RoleId = role_id
	}else if err != nil {
		return m,err
	}else if m.RoleId == conf.BookFounder{
		return m,errors.New("不能变更创始人的权限")
	}
	m.RoleId = role_id

	_,err = o.InsertOrUpdate(m)

	return m,err

}

func (m *Relationship) FindForRoleId(book_id ,member_id int) (int,error) {
	o := orm.NewOrm()

	relationship := NewRelationship()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id",book_id).Filter("member_id",member_id).One(relationship)

	if err != nil {

		return 0,err
	}
	return relationship.RoleId,nil
}

func (m *Relationship) FindByBookIdAndMemberId(book_id ,member_id int) (*Relationship,error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id",book_id).Filter("member_id",member_id).One(m)

	return m,err
}

func (m *Relationship) Insert() error  {
	o := orm.NewOrm()

	_,err :=  o.Insert(m)

	return err
}

func (m *Relationship) Update() error  {
	o := orm.NewOrm()

	_,err :=  o.Update(m)

	return err
}

func (m *Relationship) DeleteByBookIdAndMemberId(book_id,member_id int) error {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id",book_id).Filter("member_id",member_id).One(m)

	if err == orm.ErrNoRows {
		return errors.New("用户未参与该项目")
	}
	if m.RoleId == conf.BookFounder {
		return errors.New("不能删除创始人")
	}
	_,err = o.Delete(m)

	if err != nil {
		logs.Error("删除项目参与者 => ",err)
		return errors.New("删除失败")
	}
	return nil

}

func (m *Relationship) Transfer(book_id,founder_id,receive_id int) error {
	o := orm.NewOrm()

	founder := NewRelationship()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id",book_id).Filter("member_id",founder_id).One(founder)

	if err != nil {
		return err
	}
	if founder.RoleId != conf.BookFounder {
		return errors.New("转让者不是创始人")
	}
	receive := NewRelationship()

	err = o.QueryTable(m.TableNameWithPrefix()).Filter("book_id",book_id).Filter("member_id",receive_id).One(receive)

	if err != orm.ErrNoRows && err != nil {
		return err
	}
	o.Begin()

	founder.RoleId = conf.BookAdmin

	receive.MemberId = receive_id
	receive.RoleId = conf.BookFounder
	receive.BookId = book_id

	if err := founder.Update();err != nil {
		o.Rollback()
		return err
	}
	if _,err := o.InsertOrUpdate(receive);err != nil {
		o.Rollback()
		return err
	}
	return o.Commit()
}





















