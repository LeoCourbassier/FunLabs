package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/wcharczuk/go-chart"
)

var result DiscreteIntegral

// DiscreteIntegral is the struct that holds information about the integral
type DiscreteIntegral struct {
	Y         []float64
	X         []float64
	Function  string
	Result    float64
	Intervals int
	InitialX  float64
	EndX      float64
}

func (di DiscreteIntegral) f(n float64) float64 {
	ast, err := parser.ParseExpr(di.Function)
	if err != nil {
		log.Fatal(err)
	}

	return float64(Eval(ast, n))
}

// Calculate calculates the integral
func (di *DiscreteIntegral) Calculate() float64 {
	a := di.InitialX
	b := di.EndX
	coefficient := (b - a) / float64(di.Intervals)
	startingPoint := di.f(a) / 2
	endingPoint := di.f(b) / 2
	sum := 0.0
	result.Add(a, startingPoint*2)

	for i := 1; i <= di.Intervals-1; i++ {
		sum += di.f(a + (float64(i) * coefficient))
		result.Add(a+(float64(i)*coefficient), di.f(a+(float64(i)*coefficient)))
	}

	result.Add(b, endingPoint*2)

	r := coefficient * (startingPoint + sum + endingPoint)
	di.Result = r

	return r
}

// Add adds to the slice a new value calculated
func (di *DiscreteIntegral) Add(x, y float64) {
	di.X = append(di.X, x)
	di.Y = append(di.Y, y)
}

func getMin(a []float64) float64 {
	min := a[0]
	for _, e := range a {
		if e < min {
			min = e
		}
	}

	return min
}

func getMax(a []float64) float64 {
	max := a[0]
	for _, e := range a {
		if e > max {
			max = e
		}
	}

	return max
}

// Series return a chart series
func (di DiscreteIntegral) Series() chart.ContinuousSeries {
	return chart.ContinuousSeries{
		XValues: di.X,
		YValues: di.Y,
		Style: chart.Style{
			StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
			FillColor:   chart.GetDefaultColor(1).WithAlpha(64),
			Show:        true,
		},
	}
}

// GetChart returns a chart
func (di DiscreteIntegral) GetChart() chart.Chart {
	return chart.Chart{
		Title:      result.GetTitle(),
		TitleStyle: chart.StyleShow(),
		XAxis: chart.XAxis{
			Name:      "x",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			Range: &chart.ContinuousRange{
				Min: getMin(result.X),
				Max: getMax(result.X),
			},
		},
		YAxis: chart.YAxis{
			Name:      "f(x)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			Range: &chart.ContinuousRange{
				Min: getMin(result.Y),
				Max: getMax(result.Y),
			},
		},
		Series: []chart.Series{
			result.Series(),
		},
	}
}

// GetTitle returns the chart's title
func (di DiscreteIntegral) GetTitle() string {
	return fmt.Sprintf("Integral of %s from %.2f to %.2f = %.4f", strings.ReplaceAll(di.Function, " ", ""), di.InitialX, di.EndX, di.Result)
}

func main() {
	result = DiscreteIntegral{}
	result.Intervals = 1000
	result.Function = "x / (1 + (x^2))"
	result.InitialX = -5
	result.EndX = 5

	integral := result.Calculate()
	fmt.Println(integral)

	f, _ := os.Create("output.png")
	defer f.Close()
	err := result.GetChart().Render(chart.PNG, f)
	if err != nil {
		log.Fatal(err)
	}
}

// Eval evaluates an expression
func Eval(exp ast.Expr, n float64) float64 {
	switch exp := exp.(type) {
	case *ast.UnaryExpr:
		return EvalUnaryExp(exp, n)
	case *ast.BinaryExpr:
		return EvalBinaryExpr(exp, n)
	case *ast.BasicLit:
		switch exp.Kind {
		case token.INT:
			i, _ := strconv.Atoi(exp.Value)
			return float64(i)
		case token.FLOAT:
			i, _ := strconv.ParseFloat(exp.Value, 64)
			return i
		}
	case *ast.ParenExpr:
		return Eval(exp.X, n)
	case *ast.Ident:
		name := exp.Name
		switch name {
		case "e":
			return math.E
		}
		return n
	case *ast.CallExpr:
		name := exp.Fun.(*ast.Ident).Name
		arg := exp.Args[0]
		switch name {
		case "log":
			if len(exp.Args) == 2 {
				return Log(Eval(arg, n), Eval(exp.Args[1], n))
			}
			return math.Log(Eval(arg, n))
		case "sin", "sen":
			return math.Sin(Eval(arg, n))
		case "cos":
			return math.Cos(Eval(arg, n))
		case "tan":
			return math.Tan(Eval(arg, n))
		case "arcsin", "arcsen":
			return math.Asin(Eval(arg, n))
		case "arccos":
			return math.Acos(Eval(arg, n))
		case "arctan":
			return math.Atan(Eval(arg, n))
		case "mod", "abs":
			return math.Abs(Eval(arg, n))
		}

		return n
	}

	return 0
}

// Log returns the log of x in the base y
func Log(x, y float64) float64 {
	return math.Log(x) / math.Log(y)
}

// EvalBinaryExpr evaluates a binary expression
func EvalBinaryExpr(exp *ast.BinaryExpr, n float64) float64 {
	left := Eval(exp.X, n)
	right := Eval(exp.Y, n)

	switch exp.Op {
	case token.ADD:
		return left + right
	case token.SUB:
		return left - right
	case token.MUL:
		return left * right
	case token.QUO:
		return left / right
	case token.XOR:
		return math.Pow(left, right)
	}

	return 0
}

// EvalUnaryExp evaluates an unary expression
func EvalUnaryExp(exp *ast.UnaryExpr, n float64) float64 {
	arg := Eval(exp.X, n)

	switch exp.Op {
	case token.SUB:
		return -arg
	}

	return arg
}
