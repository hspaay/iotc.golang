package nodes

import "fmt"

// Unit constants with unit names.
// These are defined with the sensor type
type Unit string

// Unit name constants following ISO standard naming
const (
	UnitNone            Unit = ""
	UnitAmp             Unit = "A"
	UnitCelcius         Unit = "C"
	UnitCandela         Unit = "cd"
	UnitCount           Unit = "#"
	UnitDegree          Unit = "Degree"
	UnitFahrenheit      Unit = "F"
	UnitFeet            Unit = "ft"
	UnitGallon          Unit = "Gal"
	UnitJpeg            Unit = "jpeg"
	UnitKelvin          Unit = "K"
	UnitKmPerHour       Unit = "Kph"
	UnitLiter           Unit = "L"
	UnitMercury         Unit = "hg"
	UnitMeter           Unit = "m"
	UnitMetersPerSecond Unit = "m/s"
	UnitMilesPerHour    Unit = "mph"
	UnitMillibar        Unit = "mbar"
	UnitMole            Unit = "mol"
	UnitPartsPerMillion Unit = "ppm"
	UnitPng             Unit = "png"
	UnitKWH             Unit = "KWh"
	UnitKG              Unit = "kg"
	UnitLux             Unit = "lux"
	UnitPascal          Unit = "Pa"
	UnitPercent         Unit = "%"
	UnitPounds          Unit = "lbs"
	UnitSpeed           Unit = "m/s"
	UnitPSI             Unit = "psi"
	UnitSecond          Unit = "s "
	UnitVolt            Unit = "V"
	UnitWatt            Unit = "W"
)

// Various values for each unit
var (
	// UnitValuesAtmosphericPressure unit values for atmospheric pressure
	UnitValuesAtmosphericPressure = fmt.Sprintf("%s, %s, %s", UnitMillibar, UnitMercury, UnitPSI)
	UnitValuesImage               = fmt.Sprintf("%s, %s", UnitJpeg, UnitPng)
	UnitValuesLength              = fmt.Sprintf("%s, %s", UnitMeter, UnitFeet)
	UnitValuesSpeed               = fmt.Sprintf("%s, %s, %s", UnitMetersPerSecond, UnitKmPerHour, UnitMilesPerHour)
	UnitValuesTemperature         = fmt.Sprintf("%s, %s", UnitCelcius, UnitFahrenheit)
	UnitValuesWeight              = fmt.Sprintf("%s, %s", UnitKG, UnitPounds)
	UnitValuesVolume              = fmt.Sprintf("%s, %s", UnitLiter, UnitGallon)
)
