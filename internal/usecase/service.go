package usecase

import (
	"context"
	"time"

	"cdek_wishlist/internal/domain"

	"github.com/google/uuid"
)

type Dependencies struct {
	Users     domain.UserRepository
	Wishlists domain.WishlistRepository
	Items     domain.ItemRepository
	Hasher    PasswordHasher
	Tokens    TokenManager
	Now       func() time.Time
}

type Service struct {
	users     domain.UserRepository
	wishlists domain.WishlistRepository
	items     domain.ItemRepository
	hasher    PasswordHasher
	tokens    TokenManager
	now       func() time.Time
}

func NewService(deps Dependencies) *Service {
	return &Service{
		users:     deps.Users,
		wishlists: deps.Wishlists,
		items:     deps.Items,
		hasher:    deps.Hasher,
		tokens:    deps.Tokens,
		now:       deps.Now,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}
	if existing, err := s.users.FindByEmail(ctx, input.Email); err == nil && existing != nil {
		return nil, ErrEmailExists
	} else if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Email:        input.Email,
		PasswordHash: hash,
		CreatedAt:    s.now(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		if err == domain.ErrConflict {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (string, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		if err == domain.ErrNotFound {
			return "", ErrUnauthorized
		}
		return "", err
	}
	if err := s.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return "", ErrUnauthorized
	}
	return s.tokens.SignUserToken(user.ID)
}

func (s *Service) CreateWishlist(ctx context.Context, userID int64, input WishlistInput) (*domain.Wishlist, error) {
	if input.Name == "" {
		return nil, ErrInvalidInput
	}
	wishlist := &domain.Wishlist{
		UserID:      userID,
		Name:        input.Name,
		Description: input.Description,
		EventDate:   input.EventDate,
		Token:       uuid.NewString(),
		CreatedAt:   s.now(),
	}
	if err := s.wishlists.Create(ctx, wishlist); err != nil {
		return nil, err
	}
	return wishlist, nil
}

func (s *Service) ListWishlists(ctx context.Context, userID int64) ([]domain.Wishlist, error) {
	return s.wishlists.ListByUserID(ctx, userID)
}

func (s *Service) GetWishlist(ctx context.Context, userID, wishlistID int64) (*domain.Wishlist, error) {
	wishlist, err := s.wishlists.FindByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist.UserID != userID {
		return nil, domain.ErrForbidden
	}
	items, err := s.items.ListByWishlistID(ctx, wishlist.ID)
	if err != nil {
		return nil, err
	}
	wishlist.Items = items
	return wishlist, nil
}

func (s *Service) UpdateWishlist(ctx context.Context, userID, wishlistID int64, input WishlistInput) (*domain.Wishlist, error) {
	wishlist, err := s.wishlists.FindByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist.UserID != userID {
		return nil, domain.ErrForbidden
	}
	wishlist.Name = input.Name
	wishlist.Description = input.Description
	wishlist.EventDate = input.EventDate
	if err := s.wishlists.Update(ctx, wishlist); err != nil {
		return nil, err
	}
	return wishlist, nil
}

func (s *Service) DeleteWishlist(ctx context.Context, userID, wishlistID int64) error {
	wishlist, err := s.wishlists.FindByID(ctx, wishlistID)
	if err != nil {
		return err
	}
	if wishlist.UserID != userID {
		return domain.ErrForbidden
	}
	return s.wishlists.Delete(ctx, wishlistID)
}

func (s *Service) CreateItem(ctx context.Context, userID, wishlistID int64, input ItemInput) (*domain.Item, error) {
	wishlist, err := s.wishlists.FindByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist.UserID != userID {
		return nil, domain.ErrForbidden
	}
	item := &domain.Item{
		WishlistID:   wishlistID,
		Name:         input.Name,
		Description:  input.Description,
		Link:         input.Link,
		DesireDegree: domain.Degree(input.Degree),
	}
	if err := s.items.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdateItem(ctx context.Context, userID, wishlistID, itemID int64, input ItemInput) (*domain.Item, error) {
	wishlist, err := s.wishlists.FindByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist.UserID != userID {
		return nil, domain.ErrForbidden
	}
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.WishlistID != wishlistID {
		return nil, domain.ErrNotFound
	}
	item.Name = input.Name
	item.Description = input.Description
	item.Link = input.Link
	item.DesireDegree = domain.Degree(input.Degree)
	if err := s.items.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteItem(ctx context.Context, userID, wishlistID, itemID int64) error {
	wishlist, err := s.wishlists.FindByID(ctx, wishlistID)
	if err != nil {
		return err
	}
	if wishlist.UserID != userID {
		return domain.ErrForbidden
	}
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item.WishlistID != wishlistID {
		return domain.ErrNotFound
	}
	return s.items.Delete(ctx, itemID)
}

func (s *Service) GetPublicWishlist(ctx context.Context, token string) (*domain.Wishlist, error) {
	wishlist, err := s.wishlists.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	items, err := s.items.ListByWishlistID(ctx, wishlist.ID)
	if err != nil {
		return nil, err
	}
	wishlist.Items = items
	return wishlist, nil
}

func (s *Service) BookItem(ctx context.Context, token string, itemID int64) error {
	wishlist, err := s.wishlists.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item.WishlistID != wishlist.ID {
		return domain.ErrNotFound
	}
	if item.Booked {
		return domain.ErrAlreadyBooked
	}
	if err := s.items.Book(ctx, itemID); err != nil {
		return err
	}
	return nil
}
