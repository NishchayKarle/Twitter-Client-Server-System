// grader.go is a general purpose program that will provide students with the grade for the automated tests for assignments

package main

import (
	"encoding/json"
	"flag"
	"log"
	"os/exec"
	"strings"

	//"encoding/json"
	"fmt"
	"os"
)

type TestEvent struct {
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}
type GradescopeTest struct {
	Name     string  `json:"name"`
	Score    float32 `json:"score"`
	MaxScore float32 `json:"max_score"`
}
type Gradescope struct {
	Score      float32          `json:"score"`
	Output     string           `json:"output"`
	Visibility string           `json:"visibility"`
	StdOutVis  string           `json:"stdout_visibility"`
	Tests      []GradescopeTest `json:"tests"`
	Timeout    int              `json:"execution_time"`
}
type RubricTest struct {
	Name        string `json:"name"`        // Test name
	TestTy      string `json:"type"`        // Is this a benchmark or a normal Test?
	PackageDir  string `json:"dir"`         // The package directory where the test lives
	TableDriven bool   `json:"tabledriven"` // Is this a table-driven test
	Repeat      int    `json:"repeat"`      // Number of times to repeat the test
	Points      int    `json:"points"`      // The number of points this test is worth
	Timeout     int    `json:"timeout"`     // Overall timeout for the tests
	Total       int    `json:"total"`       // The total number of tests for this test
}

type RubricItem struct {
	Name  string       `json:"item"`  // Rubric Item name
	Tests []RubricTest `json:"tests"` // Rubric Item tests
}

type Rubric struct {
	Name    string       `json:"name"`    // Rubric file name
	Total   int          `json:"total"`   // Total points for the assignment
	Timeout int          `json:"timeout"` // Total timeout for the tests.
	Items   []RubricItem `json:"items"`   // Rubric grade items

}

func makeGradescope() {
	var gs Gradescope
	gs.Score = 0.0
	gs.Output = "We were unable to run the tests due to an error in your code."
	gs.Visibility = "visible"
	gs.StdOutVis = "visible"

	writeGradescope(&gs)
}
func writeGradescope(gs *Gradescope) {

	file, err := os.OpenFile("results.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open results.json")
		os.Exit(1)
	}

	dec := json.NewEncoder(file)

	if err := dec.Encode(gs); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if err := file.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not close file: results.json\n")
		os.Exit(1)
	}
}
func runTests(rubric Rubric, gradescopePtr *bool, gradescopeVisPtr *string) {

	var failed int
	var passed int
	var score float32
	var totalPoints float32
	var total float32
	var pScore float32
	var totalTests int
	var spawnTest bool
	var gs Gradescope
	gs.Tests = make([]GradescopeTest, len(rubric.Items))

	if !(*gradescopePtr) {
		fmt.Printf("%s\n\n", rubric.Name)
		fmt.Printf("%-62s %-6s / %-10s  %-6s / %-10s\n", "Category", "Passed", "Total", "Score", "Points")
		fmt.Println(strings.Repeat("-", 100))
	}

	for iIdx, item := range rubric.Items {
		for _, test := range item.Tests {

			for i := 0; i < test.Repeat; i++ {
				spawnTest = true
				cmd := exec.Command("go", "test", test.PackageDir, "-run", test.Name, "-json", "-count=1")
				stdout, err := cmd.StdoutPipe()

				if err != nil {
					failed = test.Total
					spawnTest = false
				}
				if err := cmd.Start(); err != nil {
					failed = test.Total
					spawnTest = false
				}

				if spawnTest {

					go func() {
						decoder := json.NewDecoder(stdout)
						for {
							var testEvent TestEvent
							if err := decoder.Decode(&testEvent); err != nil {
								break
							}
							// Ignore the last "pass" or "fail" if this a table-driven test because
							// it just states whether all tests passed or failed for that particular test.
							if test.TableDriven && testEvent.Test == fmt.Sprintf(test.Name) {
								continue
							}
							if testEvent.Action == "pass" && testEvent.Test != "" {
								passed += 1
							} else if testEvent.Action == "fail" && testEvent.Test != "" {
								failed++
							}
						}
					}()
					cmd.Wait()
				}
			}
			if failed > 0 {
				passed = 0
			} else {
				passed = 1
			}
			totalTests += test.Total
			totalPoints += float32(test.Points)
			total += float32(test.Points)
		}
		score = (float32(passed) / float32(totalTests)) * totalPoints
		pScore += score
		if !(*gradescopePtr) {
			fmt.Printf("%-62s %-6d / %-10d  %-6.2f / %-10.2f\n", item.Name, passed, totalTests, score, totalPoints)
		} else {
			var gsTest GradescopeTest
			gsTest.Name = item.Name
			gsTest.Score = score
			gsTest.MaxScore = totalPoints
			gs.Tests[iIdx] = gsTest
		}
		totalTests = 0
		totalPoints = 0
		failed = 0
		passed = 0
	}
	if total != float32(rubric.Total) {
		panic("Error: total score is not equal to total points. If you see this message then notify your instructor. ")
	}

	if !(*gradescopePtr) {
		fmt.Printf("%s\n", strings.Repeat("-", 100))
		fmt.Printf("%81s = %-6.2f / %-10f\n", "TOTAL", pScore, total)
		fmt.Println(strings.Repeat("=", 100))
		fmt.Println()
	} else {
		gs.Score = pScore
		gs.Visibility = *gradescopeVisPtr
		gs.StdOutVis = *gradescopeVisPtr
		gs.Timeout = rubric.Timeout
		writeGradescope(&gs)
	}
}

// printUsage prints the usage statement for the program
func printUsage() {
	flag.Usage()
	os.Exit(1)
}
func jsonTests(testFile string) Rubric {

	file, err := os.Open(testFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s\n", testFile)
		os.Exit(1)
	}

	dec := json.NewDecoder(file)
	var rubric Rubric

	if err := dec.Decode(&rubric); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if err := file.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not close file: %s\n", testFile)
		os.Exit(1)
	}
	return rubric
}

func main() {

	//Setup flag and positional arguments
	var assignId string
	var testFile string
	gradescopePtr := flag.Bool("gradescope", false, "Output should be in a format for gradescope")
	gradescopeVisPtr := flag.String("gradescope-visibility", "after_published", "Provide the test visibility for gradescope")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: grader ASSIGNMENT_ID\n")
		fmt.Fprintf(os.Stderr, "ASSIGNMENT_ID= the assignment identifier (e.g., hw1, proj1, etc.)\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) != 1 {
		printUsage()
	} else {
		assignId = flag.Arg(0)
		testFile = fmt.Sprintf("%s-tests.json", assignId)
		if _, err := os.Stat(testFile); err != nil {
			if os.IsNotExist(err) {
				if *gradescopePtr {
					makeGradescope()
				}
				fmt.Fprintf(os.Stderr, "Could not find rubric json test file:$s\nCheck with the instructor if you see this error.", testFile)
				printUsage()
			}
		}
		rubric := jsonTests(testFile)
		runTests(rubric, gradescopePtr, gradescopeVisPtr)
	}
}
