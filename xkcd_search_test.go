package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestParseInput(t *testing.T) {
	cmd := []string{
		"xkcd",
		"-u=false",
		"-kw",
		"hello world",
	}
	os.Args = cmd
	parseInput()
	// fmt.Println(update)
	// fmt.Println(keywords)
	if update != false && keywords != "hello world" {
		t.Fatalf("ParseInput()=%t %s want false hello world", update, keywords)
	}
}

func TestScanScriptLinesOneScript(t *testing.T) {
	input := "[[The barrel is shown, floating sideways in a choppy sea. The boy can not be seen]]\n{{title text: :( }}"
	scan := bufio.NewScanner(strings.NewReader(input))
	scan.Split(scanScriptLines)

	for scan.Scan() {
		text := scan.Text()
		want := "The barrel is shown, floating sideways in a choppy sea. The boy can not be seen"
		if text != want {
			t.Fatalf("scanScriptLines = %s want %s", text, want)
		}
	}
}

func TestScanScriptLinesMoreScripts(t *testing.T) {
	input := "[[Someone is in bed, presumably trying to sleep. The top of each panel is a thought bubble showing sheep leaping over a fence.]]\n1 ... 2 ...\n<<baaa>>\n[[Two sheep are jumping from left to right.]]\n\n... 1,306 ... 1,307 ...\n<<baaa>>\n[[Two sheep are jumping from left to right. The would-be sleeper is holding his pillow.]]\n\n... 32,767 ... -32,768 ...\n<<baaa>> <<baaa>> <<baaa>> <<baaa>> <<baaa>>\n[[A whole flock of sheep is jumping over the fence from right to left. The would-be sleeper is sitting up.]]\nSleeper: ?\n\n... -32,767 ... -32,766 ...\n<<baaa>>\n[[Two sheep are jumping from left to right. The would-be sleeper is holding his pillow over his head.]]\n\n{{Title text: If androids someday DO dream of electric sheep, don't forget to declare sheepCount as a long int.}}"
	scan := bufio.NewScanner(strings.NewReader(input))
	scan.Split(scanScriptLines)
	want1 := "Someone is in bed, presumably trying to sleep. The top of each panel is a thought bubble showing sheep leaping over a fence."
	want2 := "Two sheep are jumping from left to right."
	want3 := "Two sheep are jumping from left to right. The would-be sleeper is holding his pillow."
	want4 := "A whole flock of sheep is jumping over the fence from right to left. The would-be sleeper is sitting up."
	want5 := "Two sheep are jumping from left to right. The would-be sleeper is holding his pillow over his head."

	count := 1
	for scan.Scan() {
		text := scan.Text()
		if count == 1 && text != want1 {
			t.Fatalf("scanScriptLines %d round = %q want %q", count, text, want1)
		}

		if count == 2 && text != want2 {
			t.Fatalf("scanScriptLines %d round = %q want %q", count, text, want2)
		}

		if count == 3 && text != want3 {
			t.Fatalf("scanScriptLines %d round = %q want %q", count, text, want3)
		}

		if count == 4 && text != want4 {
			t.Fatalf("scanScriptLines %d round = %q want %q", count, text, want4)
		}

		if count == 5 && text != want5 {
			t.Fatalf("scanScriptLines %d round = %q want %q", count, text, want5)
		}
		count++
	}
}

func TestWriteIndexIntoFiles(t *testing.T) {
	//先写入
	indexMap := map[string][]int{
		"word":  {1, 2, 3, 4, 5},
		"hello": {1, 2, 3},
	}
	jsonString, _ := json.Marshal(indexMap)
	err := writeIndexIntoFiles(30, string(jsonString))
	if err != nil {
		t.Fatalf("writeIndexIntoFiles(%d %q) = %v want nil", 30, jsonString, err)
	}

	//再读取
	data, err := os.ReadFile(INDEXNUM_FNAME)
	if err != nil {
		t.Fatalf("writeIndexIntoFiles(%d %q) os.ReadFile(%v) = %v want nil", 30, jsonString, INDEXNUM_FNAME, err)
	}

	if indexNum, _ := strconv.Atoi(string(data)); indexNum != 30 {
		t.Fatalf("writeIndexIntoFiles(%d %q) = %d want 30", 30, jsonString, indexNum)
	}

	data, err = os.ReadFile(REALINDEX_FNAME)
	if err != nil {
		t.Fatalf("writeIndexIntoFiles(%d %q) os.ReadFile(%v) = %v want nil", 30, jsonString, REALINDEX_FNAME, err)
	}

	if string(data) != string(jsonString) {
		t.Fatalf("writeIndexIntoFiles(%d %q) = %q want %q", 30, jsonString, string(data), jsonString)
	}
}

func TestDoRealSearch(t *testing.T) {
	arr1 := []int{1, 2, 3, 4, 5, 6, 7, 8}
	arr2 := []int{2, 3, 4}
	r := doRealSearch(arr1, arr2)
	if !(r[0] == 2 && r[1] == 3 && r[2] == 4) {
		t.Fatalf("doRealSearch(%v %v) = %v want []int{2,3,4}", arr1, arr2, r)
	}
}

func TestDoRealSearchWithNoMatch(t *testing.T) {
	arr1 := []int{1, 2, 3, 4, 5, 6, 7, 8}
	arr2 := []int{9, 10, 11}
	r := doRealSearch(arr1, arr2)
	if len(r) != 0 {
		t.Fatalf("doRealSearch(%v %v) = %v want []int{}", arr1, arr2, r)
	}
}

func TestDoSearchOnlyOneWord(t *testing.T) {
	keywords = "hello"
	inputArr := []int{1, 2, 3, 4}
	indexmap = map[string][]int{
		"hello": inputArr,
	}
	r := doSearch()
	for i := 0; i < 4; i++ {
		if r[i] != inputArr[i] {
			t.Fatalf("doSearch(%q %#v) = %v want []int{1,2,3,4}", keywords, indexmap, r)
		}
	}
}

func TestDoSearchOnlyOneWordNoMatch(t *testing.T) {
	keywords = "hello"
	inputArr := []int{1, 2, 3, 4}
	indexmap = map[string][]int{
		"world": inputArr,
	}
	r := doSearch()
	if len(r) != 0 {
		t.Fatalf("doSearch(%q %#v) = %v want len 0 arr", keywords, indexmap, r)
	}
}

func TestDoSearchTwoWordsHaveMatch(t *testing.T) {
	keywords = "hello world"
	inputArr1 := []int{1, 2, 3, 4}
	inputArr2 := []int{1, 2, 3, 4}
	indexmap = map[string][]int{
		"hello": inputArr1,
		"world": inputArr2,
	}
	r := doSearch()
	for i := 0; i < 4; i++ {
		if r[i] != inputArr1[i] {
			t.Fatalf("doSearch(%q %#v) = %v want []int{1,2,3,4}", keywords, indexmap, r)
		}
	}
}

func TestDoSearchTwoWordsNoMatch(t *testing.T) {
	keywords = "hello world"
	inputArr1 := []int{1, 2, 3, 4}
	inputArr2 := []int{5, 6, 7, 8}
	indexmap = map[string][]int{
		"hello": inputArr1,
		"world": inputArr2,
	}
	r := doSearch()
	if len(r) != 0 {
		t.Fatalf("doSearch(%q %#v) = %v want len 0 arr", keywords, indexmap, r)
	}
}

func TestDoSearchMoreWordsHaveMatch(t *testing.T) {
	keywords = "hello world say"
	inputArr1 := []int{1, 2, 3, 4}
	inputArr2 := []int{1, 2, 3, 4}
	inputArr3 := []int{1, 2, 3}
	indexmap = map[string][]int{
		"hello": inputArr1,
		"world": inputArr2,
		"say":   inputArr3,
	}
	r := doSearch()
	for i := 0; i < 3; i++ {
		if r[i] != inputArr3[i] {
			t.Fatalf("doSearch(%q %#v) = %v want []int{1,2,3}", keywords, indexmap, r)
		}
	}
}

// func TestShowToUserWithLessThan5Results(t *testing.T) {
// 	input := []int{1, 2, 3, 4}
// 	r := showToUser(input)
// 	want := []string{
// 		"Search Result:\n",
// 		"1: https://xkcd.com/1/info.0.json\n",
// 		"2: https://xkcd.com/2/info.0.json\n",
// 		"3: https://xkcd.com/3/info.0.json\n",
// 		"4: https://xkcd.com/4/info.0.json\n",
// 		"print finished bye~\n",
// 	}

// 	if len(r) != len(want) {
// 		t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 	}

// 	for i := 0; i < len(r); i++ {
// 		if r[i] != want[i] {
// 			t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 		}
// 	}
// }

// func TestShowToUserWith5Results(t *testing.T) {
// 	input := []int{1, 2, 3, 4, 5}
// 	r := showToUser(input)
// 	want := []string{
// 		"Search Result:\n",
// 		"1: https://xkcd.com/1/info.0.json\n",
// 		"2: https://xkcd.com/2/info.0.json\n",
// 		"3: https://xkcd.com/3/info.0.json\n",
// 		"4: https://xkcd.com/4/info.0.json\n",
// 		"5: https://xkcd.com/5/info.0.json\n",
// 		"print finished bye~\n",
// 	}

// 	if len(r) != len(want) {
// 		t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 	}

// 	for i := 0; i < len(r); i++ {
// 		if r[i] != want[i] {
// 			t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 		}
// 	}
// }

// func TestShowToUserWithMoreThan5ResultsWithY(t *testing.T) {
// 	input := []int{1, 2, 3, 4, 5, 6}
// 	os.Stdin, _ = os.Open("tmpY.txt")
// 	r := showToUser(input)
// 	want := []string{
// 		"Search Result:\n",
// 		"1: https://xkcd.com/1/info.0.json\n",
// 		"2: https://xkcd.com/2/info.0.json\n",
// 		"3: https://xkcd.com/3/info.0.json\n",
// 		"4: https://xkcd.com/4/info.0.json\n",
// 		"5: https://xkcd.com/5/info.0.json\n",
// 		"6: https://xkcd.com/6/info.0.json\n",
// 		"print finished bye~\n",
// 	}

// 	if len(r) != len(want) {
// 		t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 	}

// 	for i := 0; i < len(r); i++ {
// 		if r[i] != want[i] {
// 			t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 		}
// 	}
// }

// func TestShowToUserWithMoreThan5ResultsWithN(t *testing.T) {
// 	input := []int{1, 2, 3, 4, 5, 6}
// 	os.Stdin, _ = os.Open("tmpN.txt")
// 	r := showToUser(input)
// 	want := []string{
// 		"Search Result:\n",
// 		"1: https://xkcd.com/1/info.0.json\n",
// 		"2: https://xkcd.com/2/info.0.json\n",
// 		"3: https://xkcd.com/3/info.0.json\n",
// 		"4: https://xkcd.com/4/info.0.json\n",
// 		"5: https://xkcd.com/5/info.0.json\n",
// 		"bye~",
// 	}

// 	if len(r) != len(want) {
// 		t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 	}

// 	for i := 0; i < len(r); i++ {
// 		if r[i] != want[i] {
// 			t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 		}
// 	}
// }

// func TestShowToUserWithMoreThan5ResultsWithOthers(t *testing.T) {
// 	input := []int{1, 2, 3, 4, 5, 6}
// 	os.Stdin, _ = os.Open("tmpO.txt")
// 	r := showToUser(input)
// 	want := []string{
// 		"Search Result:\n",
// 		"1: https://xkcd.com/1/info.0.json\n",
// 		"2: https://xkcd.com/2/info.0.json\n",
// 		"3: https://xkcd.com/3/info.0.json\n",
// 		"4: https://xkcd.com/4/info.0.json\n",
// 		"5: https://xkcd.com/5/info.0.json\n",
// 		"wrong param,bye~",
// 	}

// 	if len(r) != len(want) {
// 		t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 	}

// 	for i := 0; i < len(r); i++ {
// 		if r[i] != want[i] {
// 			t.Fatalf("showToUser(%v) = %v want %v", input, r, want)
// 		}
// 	}
// }
