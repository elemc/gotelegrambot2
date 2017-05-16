// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
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
