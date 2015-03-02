package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func outputCSV(entries []entry) {
	outputFile := "cs-reboot-info-output.csv"
	csvFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating csv file:", err)
		return
	}
	fmt.Printf("Writing output to %s\n", outputFile)
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	err = writer.Write([]string{"generation", "region", "server_uuid", "server_name", "reboot_window_start_UTC", "reboot_window_end_UTC", "reboot_window_start_local", "reboot_window_end_local"})
	if err != nil {
		fmt.Println("Error writing header to CSV: ", err)
	}
	for _, e := range entries {
		region := e.Region
		if e.GenType == "First Gen" {
			region = ""
		}
		record := []string{e.GenType, region, e.Server.ID, e.Server.Name, e.WindowStart.String(), e.WindowEnd.String(), e.WindowStart.Local().String(), e.WindowEnd.Local().String()}
		err := writer.Write(record)
		if err != nil {
			fmt.Println("Error writing row to CSV: ", err)
			return
		}
	}
	writer.Flush()
}
