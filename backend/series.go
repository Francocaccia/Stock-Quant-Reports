package main

func ResampleToWeekly(data []Bar) []Bar {
	if len(data) == 0 {
		return data
	}

	var result []Bar
	currentKey := -1
	var weekBar Bar

	for i, d := range data {
		year, week := d.T.ISOWeek()
		key := year*100 + week

		if i == 0 || key != currentKey {
			if i != 0 {
				result = append(result, weekBar)
			}
			currentKey = key
			weekBar = d
			continue
		}

		if d.H > weekBar.H {
			weekBar.H = d.H
		}
		if d.L < weekBar.L {
			weekBar.L = d.L
		}

		weekBar.C = d.C
		weekBar.V += d.V
	}

	result = append(result, weekBar)
	return result
}

func ResampleToMonthly(data []Bar) []Bar {
	if len(data) == 0 {
		return data
	}

	var result []Bar
	currentKey := -1
	var monthBar Bar

	for i, d := range data {
		year := d.T.Year()
		month := int(d.T.Month())
		key := year*100 + month

		if i == 0 || key != currentKey {
			if i != 0 {
				result = append(result, monthBar)
			}
			currentKey = key
			monthBar = d
			continue
		}

		if d.H > monthBar.H {
			monthBar.H = d.H
		}
		if d.L < monthBar.L {
			monthBar.L = d.L
		}

		monthBar.C = d.C
		monthBar.V += d.V
	}

	result = append(result, monthBar)
	return result
}