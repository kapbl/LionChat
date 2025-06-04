package model

type Group struct {
	ID          int        `gorm:"primaryKey;autoIncrement"`
	Name        string     `gorm:"not null"`
	Description string     `gorm:"type:text"`
	HostID      uint       `gorm:"not null;index"`                                 // 外键关联主持人[8](@ref)
	Host        *UserModel `gorm:"foreignKey:HostID;constraint:OnDelete:SET NULL"` // 主持人关联[8](@ref)
	// 组成员是一个数组，里面存放的是用户的ID
	Members []UserModel `gorm:"many2many:group_members;constraint:OnDelete:CASCADE;"` // 忽略 JSON 序列化
}
