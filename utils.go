package main

func errs2strs(errs []error) []string {
	var strs []string
	for _, v := range errs {
		strs = append(strs, v.Error())
	}
	return strs
}
