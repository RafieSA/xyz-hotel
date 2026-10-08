package service

import "xyz-hotel/backend/internal/repo"

// WishlistService handles wishlist business logic.
type WishlistService struct {
	Repo *repo.WishlistRepo
}

func NewWishlistService(r *repo.WishlistRepo) *WishlistService {
	return &WishlistService{Repo: r}
}

func (s *WishlistService) Toggle(userID, roomTypeID int64) (bool, error) {
	return s.Repo.Toggle(userID, roomTypeID)
}

func (s *WishlistService) ListEnriched(userID int64) ([]map[string]interface{}, error) {
	return s.Repo.ListEnriched(userID)
}

func (s *WishlistService) Delete(userID, roomTypeID int64) (bool, error) {
	return s.Repo.Delete(userID, roomTypeID)
}
