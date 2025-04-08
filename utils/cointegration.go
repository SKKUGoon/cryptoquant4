package utils

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// CointegrationTest tests cointegration between two time series
// It returns an approximate p-value for the ADF test on the regression residuals
func CointegrationTest(ts1, ts2 []float64) (pvalue float64, err error) {
	if len(ts1) != len(ts2) {
		return 0, fmt.Errorf("ts1 and ts2 must be the same length")
	}
	n := len(ts1)

	// Step 1: Regress ts2 on ts1 (with intercept). Compute residuals.
	beta0, beta1 := stat.LinearRegression(ts1, ts2, nil, false)
	residuals := make([]float64, n)
	for i := range n {
		predicted := beta0 + beta1*ts1[i]
		residuals[i] = ts2[i] - predicted
	}

	// Step 2: Perform ADF test on residuals
	adfStat, df, err := adfTest(residuals)
	if err != nil {
		return 0, err
	}

	tdist := distuv.StudentsT{
		Mu:    0,
		Sigma: 1,
		Nu:    float64(df),
	}
	pValue := tdist.CDF(adfStat)
	return pValue, nil
}

// adfTest performs a simple Augmented Dickey–Fuller test on a series.
// It regresses the first difference Δe[t] on a constant and lagged level e[t-1],
// returning the test statistic and degrees of freedom.
func adfTest(series []float64) (float64, int, error) {
	n := len(series)
	if n < 3 {
		return 0, 0, fmt.Errorf("series too short for ADF test")
	}

	// Compute first differences: Δe[t] = e[t] - e[t-1]
	m := n - 1
	delta := make([]float64, m)
	for t := 1; t < n; t++ {
		delta[t-1] = series[t] - series[t-1]
	}

	// Prepare regression variables:
	X := make([]float64, m)
	Y := make([]float64, m)
	for t := range m {
		X[t] = series[t]
		Y[t] = delta[t]
	}

	// Regress Y = a + γ*X using stat.LinearRegression.
	gamma, a := stat.LinearRegression(X, Y, nil, false)

	// Compute residuals and required statistics.
	residuals := make([]float64, m)
	var sumSq, meanX float64
	for _, v := range X {
		meanX += v
	}
	meanX /= float64(m)
	var sumXX float64
	for i := range m {
		yhat := a + gamma*X[i]
		residuals[i] = Y[i] - yhat
		sumSq += residuals[i] * residuals[i]
		sumXX += (X[i] - meanX) * (X[i] - meanX)
	}

	// Degrees of freedom: m - 2 (we estimate a and γ)
	df := m - 2
	if df <= 0 {
		return 0, 0, fmt.Errorf("not enough degrees of freedom for ADF test")
	}
	s2 := sumSq / float64(df)
	seGamma := math.Sqrt(s2 / sumXX)

	// t-statistic for γ.
	tStat := gamma / seGamma

	return tStat, df, nil
}
