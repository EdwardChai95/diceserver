package protocol

type OpenDiceType int

//登录状态
const (
	LoginStatusSucc = 1
	LoginStatusFail = 2
)

const (
	ActionNewAccountSignIn    = "accountLogin"
	ActionGuestSignIn         = "anonymousLogin"
	ActionOldAccountSignIn    = "oldAccountLogin"
	ActionWebChatSignIn       = "webChatSignIn"
	ActionPhoneNumberRegister = "phoneRegister"
	ActionNormalRegister      = "normalRegister"
	ActionGetVerification     = "getVerification"
	ActionAccountRegister     = "accountRegister"
)

const (
	LoginTypeAuto   = "auto"
	LoginTypeManual = "manual"
)

const (
	VerificationTypeRegister = "register"
	VerificationTypeFindPW   = "findPW"
)

// 匹配类型
const (
	MatchTypeClassic = 1 //经典
	MatchTypeDaily   = 3 //每日匹配
)

const (
	CoinTypeSliver = 0 //银币
	CoinTypeGold   = 1 //金币
)

const (
	RoomTypeClassic      = 0
	RoomTypeDailyMatch   = 1
	RoomTypeMonthlyMatch = 2
	RoomTypeFinalMatch   = 3
)

const (
	DailyMatchLevelJunior = 0
	DailyMatchLevelSenior = 1
	DailyMatchLevelMaster = 2
)

const (
	ClassicLevelJunior = 0
	ClassicLevelMiddle = 1
	ClassicLevelSenior = 2
	ClassicLevelElite  = 3
	ClassicLevelMaster = 4
)

const (
	ExitTypeExitDeskUI           = -1
	ExitTypeDissolve             = 6
	ExitTypeSelfRequest          = 0
	ExitTypeClassicCoinNotEnough = 1
	ExitTypeDailyMatchEnd        = 2
	ExitTypeNotReadyForStart     = 3
	ExitTypeChangeDesk           = 4
	ExitTypeRepeatLogin          = 5
)

const (
	DeskStatusZb      = 1
	DeskStatusDq      = 2
	DeskStatusPlaying = 3
	DeskStatusEnded   = 4
)

const (
	OpenTypeWin OpenDiceType = iota
	OpenTypeLose
)

const (
	SexTypeUnknown = 0
	SexTypeMale    = 1
	SexTypeFemale  = 2
)

const (
	UserTypeGuest    = 0
	UserTypeLaoBaShi = 1
)

// OpType
const (
	OptypeIllegal = 0
	OptypeCall    = 1
	OptypeOpen    = 2
	OptypePass    = 3
)

// 创建房间频道选项
const (
	ChannelOptionAll  = "allChannel"
	ChannelOptionHalf = "halfChannel"
)
