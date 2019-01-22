
## API

 - #####  User
 #
- - ######  `/enter = {login:string, pass:string} => {result:string, token:string}`
#
- - ######  `/testToken =  {token:string} = > {result:string}`
#
- - ######  `/create = {login:string, pass:string} => {result:string, token:string}`
#
- - ######  `/getMyChats = {token:string} => {[...{ }]}`
#
- - ######  `/myData = {login:string, pass:string} => {id: int64}`
#
- - ######  `/getSettings = {login:string, pass:string} => {login:string, name:string}`
#
- - ######  `/setSettings =  {login:string, pass:string} => {name:string}`