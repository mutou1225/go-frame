package commodel

const (
	TTChannel 	= "t_channel"
	TTMaptag 	= "t_maptag"
	TTag 		= "t_tag"
)

type TChannel struct {
	ChId     	int    	  `gorm:"primary_key;column:Fchannel_id"` 	// 渠道ID
	ChName  	string    `gorm:"column:Fchannel_name"`           	// 渠道名称
}

type TPid struct {
	Pid     	int    		`gorm:"column:Fp_id"` 			// PID
	TagId  		int    		`gorm:"column:Ftag_id"`         // TagId
	TagName    	string    	`gorm:"column:Ftag_name"` 		// tag_name
	ChId     	int    	  	`gorm:"column:Fchannel_id"` 	// 渠道ID
	ChName  	string    	`gorm:"column:Fchannel_name"`   // 渠道名称
}

