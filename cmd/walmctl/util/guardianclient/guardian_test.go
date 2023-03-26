package guardianclient

//func getTestClient() *GuardianClient {
//	guardianClient := NewClient("https://172.16.1.70/integ1/guardian/guardian", "admin", "123")
//
//	return guardianClient
//}
//
//func Test_GetUsers(t *testing.T) {
//	guardianClient := getTestClient()
//
//	users, err := guardianClient.GetUsers()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Printf("users %v\n", users)
//}
//
//func Test_GetKeytabs(t *testing.T) {
//	guardianClient := getTestClient()
//
//	principals := []string{
//		"hdfs", "hbase", "hive",
//	}
//	_, err := guardianClient.GetMultipleKeytabs(principals)
//	if err != nil {
//		panic(err)
//	}
//}
