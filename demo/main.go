package main

import (
	"fmt"
	"github/fconf"
	"regexp"
	"strings"
)

/*
函数名：delete_extra_space(s string) string
功  能:删除字符串中多余的空格(含tab)，有多个空格时，仅保留一个空格，同时将字符串中的tab换为空格
参  数:s string:原始字符串
返回值:string:删除多余空格后的字符串
创建时间:2018年12月3日
创建者:王俊鹏
修订信息:
*/
func delete_extra_space(s string) string {
	//删除字符串中的多余空格，有多个空格时，仅保留一个空格
	s1 := strings.Replace(s, "	", " ", -1)       //替换tab为空格
	regstr := "\\s{2,}"                          //两个及两个以上空格的正则表达式
	reg, _ := regexp.Compile(regstr)             //编译正则表达式
	s2 := make([]byte, len(s1))                  //定义字符数组切片
	copy(s2, s1)                                 //将字符串复制到切片
	spc_index := reg.FindStringIndex(string(s2)) //在字符串中搜索
	for len(spc_index) > 0 {                     //找到适配项
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...) //删除多余空格
		spc_index = reg.FindStringIndex(string(s2))            //继续在字符串中搜索
	}
	return string(s2)
}

//数据库字段类型结构
type db_column struct {
	name     string
	col_type string
}

/*
函数名：getDBColumnsFromStr(s string) []string
功  能:从字符串中获取以“,"分割的数据库列名称信息，列名包括名称和数据类型，以空格分割
参  数:s string:原始字符串
返回值:[]string:列明和数据类型字符串组
创建时间:2018年12月3日
创建者:王俊鹏
修订信息:
*/
func getDBColumnsFromStr(s string) []db_column {
	str := strings.Split(delete_extra_space(s), ",")
	m := make([]db_column, len(str))
	for i := 0; i < len(str); i++ {
		s1 := strings.Split(delete_extra_space(str[i]), ":")
		m[i].name = s1[0]
		m[i].col_type = s1[1]
	}
	return m
}

func main() {
	c, err := fconf.NewFileConf("./demo.ini")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(c.String("mysql.db1.Host"))
	fmt.Println(c.String("mysql.db1.Name"))
	fmt.Println(c.String("mysql.db1.User"))
	fmt.Println(c.String("mysql.db1.Pwd"))
	fmt.Println(c.String("tcp.Port"))
	fmt.Println(fconf.GetDBColumns(c.String("mysql.db1.colname")))

	// 取得配置时指定类型
	port, err := c.Int("mysql.db1.Port")
	if err != nil {
		panic(err)
	}
	fmt.Println(port) // output:127.0.0.1
}
