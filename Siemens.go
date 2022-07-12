package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type k27804 struct {
	key string
	kd  [5]kdata
}

type kdata struct {
	kLSI   string
	kBML   string
	kValue string
	kCmt   string
}

func failOnError(err error) {
	if err != nil {
		log.Fatal("Error:", err)
	}
}

func main() {
	flag.Parse()

	//ログファイル準備
	logfile, err := os.OpenFile("./log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	failOnError(err)
	defer logfile.Close()

	log.SetOutput(logfile)
	// log.Print("Start\r\n")

	// 読み込んだファイルは削除する
	defer os.Remove(flag.Arg(0))

	// ファイルの読み込み準備
	infile, err := os.Open(flag.Arg(0))
	failOnError(err)
	defer infile.Close()

	// ファイルの書き込み準備
	// writeFileDir, _ := filepath.Split(flag.Arg(0))

	// LSIFileDir := writeFileDir + "LSI.DAT"
	// LSIfile, err := os.OpenFile(LSIFileDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// failOnError(err)
	// defer LSIfile.Close()

	BMLFileDir := `\\192.168.4.22\siemens\k27804.DAT`
	BMLfile, err := os.OpenFile(BMLFileDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	failOnError(err)
	defer BMLfile.Close()

	r := csv.NewReader(transform.NewReader(infile, japanese.ShiftJIS.NewDecoder()))

	row, err := r.Read() // csvを1行ずつ読み込む 1行目はタイトルなので使用しない
	if err == io.EOF {
		failOnError(err)
	}

	var ik27804 k27804
	var LSIlines [4]string
	var BMLlines [4]string

	for {
		//ik27804の初期化
		ik27804.key = ""
		for i := 0; i < 5; i++ {
			ik27804.kd[i].kLSI = ""
			ik27804.kd[i].kBML = ""
			ik27804.kd[i].kValue = ""
			ik27804.kd[i].kCmt = ""
		}

		//LSI･BMLlinesの初期化
		for w := 0; w < 4; w++ {
			LSIlines[w] = ""
			BMLlines[w] = ""
		}

		row, err = r.Read() // csvを1行ずつ読み込む
		if err == io.EOF {
			break
		}

		ik27804.key = row[1][:4] + row[1][5:7] + row[1][8:] + fmt.Sprintf("%07s", row[2])

		w := 0
		j := 0
		for i := 8; i < 25; i++ {

			if row[i] != "" {
				ik27804.kd[j].kLSI = codeLSI(i)
				ik27804.kd[j].kBML = codeBML(i)
				ik27804.kd[j].kValue = fmt.Sprintf("% 8s", vConv(row[i]))
				ik27804.kd[j].kCmt = strings.Repeat(" ", 7)

				j++
			}

			if j >= 5 {
				LSIlines[w] = LSIset(ik27804)
				BMLlines[w] = BMLset(ik27804)
				w++
				j = 0

				//ik27804の初期化
				ik27804.key = row[1][:4] + row[1][5:7] + row[1][8:] + fmt.Sprintf("%07s", row[2])
				for i := 0; i < 5; i++ {
					ik27804.kd[i].kLSI = ""
					ik27804.kd[i].kBML = ""
					ik27804.kd[i].kValue = ""
					ik27804.kd[i].kCmt = ""
				}

			}
		}

		if ik27804.kd[0].kLSI != "" {
			LSIlines[w] = LSIset(ik27804)
			BMLlines[w] = BMLset(ik27804)
		}

		// 一人分のデータを書き込み
		for w := 0; w < 4; w++ {
			// if LSIlines[w] != "" {
			// 	_, err = LSIfile.WriteString(LSIlines[w] + "\r\n")
			// 	failOnError(err)
			// }
			if BMLlines[w] != "" {
				_, err = BMLfile.WriteString(BMLlines[w] + "\r\n")
				failOnError(err)
			}
		}

		// log.Println(row)

	}

	log.Print("データを変換しました。\r\n")

}

func LSIset(k k27804) string {
	var line string

	line = "A1"                                  // レコード区分
	line = line + strings.Repeat(" ", 6)         // センターコード
	line = line + k.key + strings.Repeat(" ", 5) // 依頼元key
	line = line + strings.Repeat(" ", 20)        // 受託者key
	line = line + strings.Repeat(" ", 20)        // 被験者名
	line = line + "E"                            // 報告状況コード
	line = line + strings.Repeat(" ", 9)         // 検体状態
	for i := 0; i < 5; i++ {
		if k.kd[i].kLSI == "" {
			line = line + strings.Repeat(" ", 32)
		} else {
			line = line + k.kd[i].kLSI + strings.Repeat(" ", 12)
			line = line + k.kd[i].kValue
			line = line + k.kd[i].kCmt
		}
	}
	line = line + strings.Repeat(" ", 18)

	return line
}

func BMLset(k k27804) string {
	var line string

	line = "A1"                                  // レコード区分
	line = line + strings.Repeat(" ", 6)         // センターコード
	line = line + k.key + strings.Repeat(" ", 5) // 依頼元key
	line = line + strings.Repeat(" ", 20)        // 受託者key
	line = line + strings.Repeat(" ", 20)        // 被験者名
	line = line + "E"                            // 報告状況コード
	line = line + strings.Repeat(" ", 9)         // 検体状態
	for i := 0; i < 5; i++ {
		if k.kd[i].kBML == "" {
			line = line + strings.Repeat(" ", 32)
		} else {
			line = line + k.kd[i].kBML + strings.Repeat(" ", 12)
			line = line + k.kd[i].kValue
			line = line + k.kd[i].kCmt
		}
	}
	line = line + strings.Repeat(" ", 18)

	return line
}

func codeLSI(v int) string {
	var code string

	switch v {
	case 8:
		code = "00605"
	case 9:
		code = "00610"
	case 10:
		code = "00653"
	case 11:
		code = "00639"
	case 12:
		code = "00638"
	case 13:
		code = "00637"
	case 14:
		code = "00601"
	case 15:
		code = "99999"
	case 16:
		code = "00609"
	case 17:
		code = "99999"
	case 18:
		code = "99999"
	case 19:
		code = "99999"
	case 20:
		code = "99999"
	case 21:
		code = "99999"
	case 22:
		code = "99999"
	case 23:
		code = "99999"
	case 24:
		code = "99999"
	case 25:
		code = "99999"
	}

	return code

}

func codeBML(v int) string {
	var code string

	switch v {
	case 8:
		code = "00055"
	case 9:
		code = "00059"
	case 10:
		code = "00062"
	case 11:
		code = "00060"
	case 12:
		code = "00063"
	case 13:
		code = "00061"
	case 14:
		code = "00051"
	case 15:
		code = "99999"
	case 16:
		code = "00057"
	case 17:
		code = "99999"
	case 18:
		code = "99999"
	case 19:
		code = "99999"
	case 20:
		code = "99999"
	case 21:
		code = "99999"
	case 22:
		code = "99999"
	case 23:
		code = "99999"
	case 24:
		code = "99999"
	case 25:
		code = "99999"
	}

	return code

}

func vConv(v string) string {
	var s string
	switch v {
	case "-":
		s = "(-)"
	case "+/-":
		s = "(+-)"
	case "+/-ﾋﾖｳｹﾂ":
		s = "(+-)"
	case "+/- 溶血":
		s = "(+-)"
	case "+":
		s = "(+)"
	case "1+":
		s = "(+)"
	case "2+":
		s = "(2+)"
	case "3+":
		s = "(3+)"
	case "4+":
		s = "(4+)"
	case "5+":
		s = "(5+)"
	case "<=1.005":
		s = "1.005"
	case ">=1.030":
		s = "1.030"
	default:
		s = v
	}

	return s
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
