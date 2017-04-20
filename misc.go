// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

func appendStringToSliceIfNotFound(slice []string, str string) []string {
	for _, l := range slice {
		if l == str {
			return slice
		}
	}

	slice = append(slice, str)
	return slice
}
