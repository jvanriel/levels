package main

/*
	Adjust
	- Adjust dBFS to dBSPL
	- Add fixed offset from options
	- Add sensitivity from calibration files
	- Add interpolated SPL from calibration files
*/

func (s *Server) adjust(channel int, dBFS float64) float64 {
	dBSPL := dBFS

	// Add fixed offset from options, default is 94.0
	// FIXME: Don't know REW's default
	dBSPL += float64(s.sploffset)

	// Add sensitivity from calibration files
	// FIXME: I'm not sure what to do with sensitivity
	dBSPL += s.calfiles.sensitivity(channel)

	// Add interpolated SPL from calibration files
	// FIXME: I'm not sure if this is correct
	dBSPL += s.calfiles.interpolatedSPL(channel)

	return dBSPL
}
