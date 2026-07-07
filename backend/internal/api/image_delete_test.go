package api

import "testing"

//Deleting image
//Delete image
//Delete image storage fails db row survives
//Delete image db fails storage object orphaned
//Delete image no ownership

func TestDeleteImage(t *testing.T) {
	// tests := []struct {
	// 	name       string
	// 	filename   string
	// 	userID     uint64
	// 	wantStatus int
	// }{
	// 	{
	// 		"delete success",
	// 		"image.png",
	// 		2,
	// 		http.StatusOK,
	// 	},
	// 	{
	// 		"delete succes: moderator user but not an owner",
	// 		"image.png",
	// 		1,
	// 		http.StatusOK,
	// 	},
	// 	{
	// 		"delete forbidden: user is not an owner or a moderator",
	// 		"image.png",
	// 		3,
	// 		http.StatusForbidden,
	// 	},
	// 	{
	// 		"delete failure: user unathorized",
	// 		"image.png",
	// 		0,
	// 		http.StatusUnauthorized,
	// 	},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		a, store := newTestAPI(t)
	// 		r := newTestRouter(a)

	// 		testImages := {}
	// 	})
	// }
}
