package fconf

/**
 * 截入配置文件并读取其中内容
 * @Auther QiuXiangCheng
 * @Date   2018/05/08
 *
 * 与INI配置文件风格一样 根据顺序读取文件和每一行 如果在行首出现了;号，则认为是配置文件的注释
 * 当INI不规范时 如[mysql] 的注释被写为[mysql 则会返回错误
 * 本包的配置文件是严格区分大小写的 需要禁止区分大小写 将在后期加入或自行加入
 */

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 配置文件
type Config struct {
	conf map[string]url.Values
}

// 将指定的配置以字符串返回
func (c *Config) String(tag string) string {
	spl := strings.Split(tag, ".")
	key := strings.Join(spl[1:], "_")
	if len(spl) < 2 || spl[1] == "" {
		return ""
	}

	return c.conf[spl[0]].Get(key)
}

// 返回一个Int类型的配置值
func (c *Config) Int(tag string) (int, error) {
	return strconv.Atoi(c.String(tag))
}

// 返回一个int64配置值
func (c *Config) Int64(tag string) (int64, error) {
	return strconv.ParseInt(c.String(tag), 10, 64)
}

// 返回一个float64配置值
func (c *Config) Float64(tag string) (float64, error) {
	return strconv.ParseFloat(c.String(tag), 64)
}

// 初始化一个文件配置句柄
func NewFileConf(filePath string) (*Config, error) {

	cf := &Config{
		conf: make(map[string]url.Values, 10),
	}

	f, err := NewFileReader(filePath)
	if err != nil {
		return nil, errors.New("Error:can not read file \"" + filePath + "\"")
	}
	defer f.Close()

	tag := ""
	buf := bufio.NewReader(f)
	replacer := strings.NewReplacer(" ", "")

	for {
		lstr, err := buf.ReadString('\n')
		if err != nil && err != errors.New("EOF") {
			break
		}

		if lstr == "" {
			break
		}

		lstr = strings.TrimSpace(lstr)
		if lstr == "" {
			continue
		}

		if idx := strings.Index(lstr, "["); idx != -1 {
			if lstr[len(lstr)-1:] != "]" {
				return nil, errors.New("Error:field to parse this symbol style:\"" + lstr + "\"")
			}
			tag = lstr[1 : len(lstr)-1]
			cf.conf[tag] = url.Values{}
		} else {
			lstr = replacer.Replace(lstr)
			spl := strings.Split(lstr, "=")

			if lstr[0:1] == ";" {
				continue
			}

			if len(spl) < 2 {
				return nil, errors.New("error:" + lstr)
			}
			cf.conf[tag].Set(strings.Replace(spl[0], ".", "_", -1), spl[1])
		}
	}

	return cf, nil
}

// 打开一个文件句柄
func NewFileReader(filePath string) (*os.File, error) {
	if !PathExists(filePath) {
		return nil, errors.New("Error:File not exists:" + filePath)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// 检查文件或文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

/*
函数名：delete_extra_space(s string) string
功  能:删除字符串中多余的空格(含tab)，有多个空格时，仅保留一个空格，同时将字符串中的tab换为空格
参  数:s string:原始字符串
返回值:string:删除多余空格后的字符串
创建时间:2018年12月3日
创建者:王俊鹏
修订信息:
*/
func deleteExtraSpace(s string) string {
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
type DbColumn struct {
	Name    string
	ColType string
}

/*
函数名：GetDBColumnsFromStr(s string) DbColumn
功  能:从字符串中获取以“,"分割的数据库列名称信息，列名包括名称和数据类型，以空格分割
参  数:s string:原始字符串
返回值:DbColumn:列名和数据类型结构数组
创建时间:2018年12月3日
创建者:王俊鹏
修订信息:
*/
func GetDBColumnsMsg(s string) []DbColumn {
	str := strings.Split(deleteExtraSpace(s), ",")
	m := make([]DbColumn, len(str))
	if len(str) > 0 {
		for i := 0; i < len(str); i++ {
			s1 := strings.Split(deleteExtraSpace(str[i]), ":")
			if len(s1) < 2 {
				m[i].Name = s1[0]
			} else {
				m[i].Name = s1[0]
				m[i].ColType = s1[1]
			}
		}
	} else {
		m = nil
	}
	return m
}

/*
函数名：GetDBColumnsStr(msg []DbColumn) string
功  能:将数据库列信息中的列名提取出来，组成一个列名数据，使其方便用于SQL语句
参  数:msg []DbColumn：包含列名、列类型的数据库列信息结构数组
返回值:string:列名字符串组
创建时间:2018年12月15日
创建者:王俊鹏
修订信息:
*/
func GetDBColumnsStr(msg []DbColumn) string {
	var str_col []byte
	for i := range msg {
		if i < len(msg)-1 {
			str_col = []byte(fmt.Sprintf("%s%s,", string(str_col), msg[i].Name))
		} else {
			str_col = []byte(fmt.Sprintf("%s%s", string(str_col), msg[i].Name))
		}
	}
	return string(str_col)
}

/*
函数名：GetCfg(tag string, cfg string)string
参  数:tag string:参数的名称
	  filepath string:文件路径
返回值:string:从配置信息文件中查询到的第一个tag参数的配置值
创建时间:2018年12月3日
修订信息:
*/
func GetCfg(filepath, tag string) (string, error) {
	dat, err := ioutil.ReadFile(filepath) //读取文件
	cfg := string(dat)                    //将读取到达配置文件转化为字符串
	var str string
	s1 := fmt.Sprintf("[^;]%s\\s*=\\s*.{1,}\\n", tag)
	s2 := fmt.Sprintf("%s\\s*=\\s*", tag)
	reg, _ := regexp.Compile(s1)
	if err == nil {
		tag_str := reg.FindString(cfg) //在配置字符串中搜索
		if len(tag_str) > 0 {
			r, _ := regexp.Compile(s2)
			i := r.FindStringIndex(tag_str) //查找配置字符串的确切起始位置
			var h_str = make([]byte, len(tag_str)-i[1])
			copy(h_str, tag_str[i[1]:])
			str1 := fmt.Sprintln(string(h_str))
			str2 := strings.Replace(str1, "\n", "", -1)
			str = strings.Replace(str2, "\r", "", -1)
		}
	}
	return str, err
}

/*
函数名：WriteTagValueToFile(filepath, tag, content string)
参  数:filepath string:文件路径
	  tag string:变量标签
	  content string:变量值
返回值:无
创建时间:2018年12月20日
修订信息:
*/
func WriteTagValueToFile(filepath, tag, content string) error {
	dat, err := ioutil.ReadFile(filepath) //读取文件
	if err != nil {
		dat = append(dat, "[MicETL]\r"...)
	}
	cfg := string(dat) //将读取到达配置文件转化为字符串
	var str, tag_str string
	s1 := fmt.Sprintf("[^;]%s\\s*=\\s*.{1,}\\n", tag)
	s2 := fmt.Sprintf("%s\\s*=\\s*", tag)
	reg, err := regexp.Compile(s1)
	if err == nil {
		tag_str = reg.FindString(cfg) //在配置字符串中搜索
		//fmt.Println("搜索结果:", tag_str, "正则表达式:", s1)
		if len(tag_str) > 0 {
			r, _ := regexp.Compile(s2)
			i := r.FindStringIndex(tag_str) //查找配置字符串的确切起始位置
			var h_str = make([]byte, len(tag_str)-i[1])
			copy(h_str, tag_str[i[1]:])
			str1 := fmt.Sprintln(string(h_str))

			str2 := strings.Replace(str1, "\n", "", -1)
			str3 := strings.Replace(str2, "\r", "", -1)
			str = strings.Replace(tag_str, str3, content, -1)
			//fmt.Println("str1=", str1, "str2=", str2, "str3=", str3, "str=", str, "tag_str=", tag_str)
		} else {
			tag_str = fmt.Sprintln(tag, "=", content)
			str = fmt.Sprint(tag_str)
			dat = append(dat, str...) //添加到最后
			cfg = string(dat)
		}
	}
	//fmt.Printf("*tag_str=%s\n", tag_str)

	fileObj, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		//fmt.Println("打开文件失败:", err.Error())
		os.Exit(2)
	}
	defer fileObj.Close()
	//fmt.Println("cfg:", cfg, "tag_str:", tag_str, "str:", str)
	contents := []byte(strings.Replace(cfg, tag_str, str, -1))
	if _, err := fileObj.Write(contents); err == nil {
		//fmt.Println("写入文件成功")
	}
	return err
}

/*
函数名：WriteLog(filepath, content string)
参  数:filepath string:文件路径,不用加后缀名,程序自动添加txt后缀名,并且自动添加日期
	  content string:需要记录的信息
返回值:无
创建时间:2018年12月20日
修订信息:
*/
func WriteLog(filepath, content string) {
	name := fmt.Sprintf("%s_%s.txt", filepath, time.Now().Format("2006-01-02"))
	if fileObj, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err == nil {
		defer fileObj.Close()
		writeObj := bufio.NewWriterSize(fileObj, 4096)

		//使用Write方法,需要使用Writer对象的Flush方法将buffer中的数据刷到磁盘
		buf := []byte(fmt.Sprintf("%s  %s\n", time.Now().Format("2006-01-02 15:04:05"), content))
		if _, err := writeObj.Write(buf); err == nil {
			if err := writeObj.Flush(); err != nil {
				panic(err)
			}
		}
	}
}
