package api

const (
	// Common
	failedDecodeData      = iota + 1
	failedGetSettings     //2
	haventRightsForAction //3
	failedEncodeData      //5
	endPointNotFound      //6
	invalidToken          //6
)

// Chat API
const (
	createdCahnnel          = iota + 1
	userChatCheckFailed     //2
	userIsDeletedFromChat   //3
	failedGetUserInfo       //4
	failedGetUsersForAdd    //5
	failedDeleteUsers       //6
	failedRecoveryUsers     //7
	failedGetChatSettings   //8
	shortChatName           //9
	failedSetChatSettings   //10
	failedDeleteFromList    //11
	createdChat             //12
	failedGetUsersForDialog //13
)

//File API
const (
	failedDecodeFromData   = iota + 1
	failedGetDataFromForm  //2
	failedRebuildDataTypes //3
	failedCreatFile        //4
	failedOpenFile         //5
	failedDeleteFileDB     //6
	failedDeleteFileOS     //7
	fileDoesntExist        //8
)

//Messages API
const (
	failedGetAdditionalMessages = iota + 1
	failedGetMessages
)

//User API
const (
	failedGetUser         = iota + 1
	failedGenerateToken   //2
	someEmptyFields       //3
	failedCreateUser      //4
	failedGetUserChats    //5
	failedSetUserSettings //6
)
