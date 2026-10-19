package app

import (
	"context"
	"math"
	"strings"
	"sync"

	"amiguis/internal/catalog/domain"
	contentApp "amiguis/internal/content/app"
	contentDomain "amiguis/internal/content/domain"
)

type ReviewInput struct {
	AuthorID string
	ItemID   string
	Body     string
	Rating   int
	Watched  bool
	Liked    bool
}

type CatalogService struct {
	mu             sync.RWMutex
	items          []*domain.CatalogItem
	contentService *contentApp.ContentService
}

func NewCatalogService(contentService *contentApp.ContentService) *CatalogService {
	return &CatalogService{
		items:          defaultCatalogItems(),
		contentService: contentService,
	}
}

func (s *CatalogService) Search(ctx context.Context, query string, category string) []*domain.CatalogItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	cat := strings.ToLower(strings.TrimSpace(category))

	var results []*domain.CatalogItem
	for _, item := range s.items {
		if cat != "" && cat != "all" && string(item.Category) != cat {
			continue
		}
		if q != "" {
			matchTitle := strings.Contains(strings.ToLower(item.Title), q)
			matchCreator := strings.Contains(strings.ToLower(item.Creator), q)
			matchProducer := strings.Contains(strings.ToLower(item.Producer), q)
			matchCast := false
			for _, member := range item.Cast {
				if strings.Contains(strings.ToLower(member), q) {
					matchCast = true
					break
				}
			}
			if !matchTitle && !matchCreator && !matchProducer && !matchCast {
				continue
			}
		}
		results = append(results, item)
	}
	if results == nil {
		return []*domain.CatalogItem{}
	}
	return results
}

func (s *CatalogService) GetByID(ctx context.Context, id string) (*domain.CatalogItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, domain.ErrCatalogItemNotFound
}

func (s *CatalogService) SubmitReview(ctx context.Context, in ReviewInput) (*contentApp.PostDTO, *domain.CatalogItem, error) {
	if in.Rating < 0 || in.Rating > 5 {
		return nil, nil, domain.ErrInvalidRating
	}

	s.mu.Lock()
	var target *domain.CatalogItem
	for _, item := range s.items {
		if item.ID == in.ItemID {
			target = item
			break
		}
	}
	if target == nil {
		s.mu.Unlock()
		return nil, nil, domain.ErrCatalogItemNotFound
	}

	if in.Rating > 0 {
		totalScore := target.AverageRating*float64(target.RatingsCount) + float64(in.Rating)
		target.RatingsCount++
		target.AverageRating = math.Round((totalScore/float64(target.RatingsCount))*10) / 10
	}
	itemCopy := *target
	s.mu.Unlock()

	reviewMeta := &contentDomain.ReviewMetadata{
		ItemID:    itemCopy.ID,
		Category:  string(itemCopy.Category),
		ItemTitle: itemCopy.Title,
		ItemYear:  itemCopy.Year,
		PosterURL: itemCopy.PosterURL,
		Rating:    in.Rating,
		Watched:   in.Watched,
		Liked:     in.Liked,
		Creator:   itemCopy.Creator,
		Producer:  itemCopy.Producer,
	}

	body := strings.TrimSpace(in.Body)
	if body == "" {
		if in.Watched {
			body = "Marcó \"" + itemCopy.Title + "\" como vista/visto."
		} else {
			body = "Compartió una valoración sobre \"" + itemCopy.Title + "\"."
		}
	}

	postDTO, err := s.contentService.CreatePost(ctx, contentApp.CreatePostInput{
		AuthorID: in.AuthorID,
		Type:     contentDomain.PostTypeReview,
		Body:     body,
		Review:   reviewMeta,
	})
	if err != nil {
		return nil, nil, err
	}

	return postDTO, &itemCopy, nil
}

func defaultCatalogItems() []*domain.CatalogItem {
	return []*domain.CatalogItem{
		{
			ID:            "cat-movie-1",
			Category:      domain.CategoryMovie,
			Title:         "El abrazo de la serpiente",
			Year:          2015,
			Description:   "La épica historia del primer contacto, encuentro, acercamiento, traición y posible amistad que trasciende la vida entre Karamakate, un chamán amazónico y último sobreviviente de su tribu, y dos científicos que recorren el Amazonas en busca de la yakruna, una planta sagrada.",
			PosterURL:     "https://images.unsplash.com/photo-1518709268805-4e9042af9f23?auto=format&fit=crop&w=600&q=80",
			AverageRating: 4.8,
			RatingsCount:  342,
			Cast:          []string{"Nilbio Torres", "Antonio Bolívar", "Jan Bijvoet", "Brionne Davis"},
			Producer:      "Ciudad Lunar Producciones (Christina Gallego)",
			Creator:       "Ciro Guerra",
			Genres:        []string{"Drama", "Aventura", "Historia"},
			Duration:      "125 min",
			Country:       "Colombia",
		},
		{
			ID:            "cat-movie-2",
			Category:      domain.CategoryMovie,
			Title:         "Encanto",
			Year:          2021,
			Description:   "En lo alto de las montañas de Colombia hay un lugar encantado llamado Encanto. Aquí, en una casa mágica, vive la extraordinaria familia Madrigal donde todos tienen habilidades fantásticas excepto Mirabel, quien descubre que la magia que rodea el hogar está en peligro.",
			PosterURL:     "https://images.unsplash.com/photo-1513151233558-d860c5398176?auto=format&fit=crop&w=600&q=80",
			AverageRating: 4.7,
			RatingsCount:  520,
			Cast:          []string{"Stephanie Beatriz", "María Cecilia Botero", "John Leguizamo", "Mauro Castillo", "Carolina Gaitán"},
			Producer:      "Walt Disney Animation Studios (Yvett Merino, Clark Spencer)",
			Creator:       "Byron Howard & Jared Bush",
			Genres:        []string{"Animación", "Musical", "Familia"},
			Duration:      "102 min",
			Country:       "Colombia / Estados Unidos",
		},
		{
			ID:            "cat-movie-3",
			Category:      domain.CategoryMovie,
			Title:         "La estrategia del caracol",
			Year:          1993,
			Description:   "Los inquilinos de una vieja casona en el centro de Bogotá, amenazados por una orden de desalojo, idean un ingenioso e inolvidable plan liderado por Don Jacinto y el Perro Romero para llevarse la casa entera pieza por pieza.",
			PosterURL:     "https://images.unsplash.com/photo-1489599849927-2ee91cede3ba?auto=format&fit=crop&w=600&q=80",
			AverageRating: 4.9,
			RatingsCount:  612,
			Cast:          []string{"Fausto Cabrera", "Frank Ramírez", "Vicky Hernández", "Carlos Vives", "Humberto Dorado"},
			Producer:      "Fotoclub 76 / Caracol Televisión",
			Creator:       "Sergio Cabrera",
			Genres:        []string{"Comedia dramática", "Cine colombiano"},
			Duration:      "116 min",
			Country:       "Colombia",
		},
		{
			ID:            "cat-series-1",
			Category:      domain.CategorySeries,
			Title:         "Cien años de soledad",
			Year:          2024,
			Description:   "Al casarse en contra de la voluntad de sus padres, los primos José Arcadio Buendía y Úrsula Iguarán dejan atrás su pueblo y emprenden un largo viaje en busca de un nuevo hogar que culmina con la fundación del mítico pueblo de Macondo.",
			PosterURL:     "https://images.unsplash.com/photo-1509198397868-475647b2a1e5?auto=format&fit=crop&w=600&q=80",
			AverageRating: 4.8,
			RatingsCount:  489,
			Cast:          []string{"Claudio Cataño", "Marco González", "Susana Morales", "Viña Machado", "Diego Vásquez"},
			Producer:      "Dynamo Producciones / Netflix (Rodrigo García, Gonzalo García Barcha)",
			Creator:       "Laura Mora & Alex García López",
			Genres:        []string{"Drama", "Realismo mágico", "Historia"},
			Duration:      "1 temporada (16 episodios)",
			Country:       "Colombia",
		},
		{
			ID:            "cat-series-2",
			Category:      domain.CategorySeries,
			Title:         "Secuestro del vuelo 601",
			Year:          2024,
			Description:   "Dos revolucionarios armados secuestran y amenazan con hacer explotar el Vuelo 601 a menos que el gobierno colombiano libere a 50 prisioneros políticos y les pague un cuantioso rescate en efectivo, mientras dos valientes azafatas luchan por mantener a salvo a los pasajeros.",
			PosterURL:     "https://images.unsplash.com/photo-1436491865332-7a61a109cc05?auto=format&fit=crop&w=600&q=80",
			AverageRating: 4.5,
			RatingsCount:  275,
			Cast:          []string{"Mónica Lopera", "Christian Tappan", "Ángela Cano", "Enrique Carriazo", "Marcela Benjumea"},
			Producer:      "Miracol Media / Netflix",
			Creator:       "Pablo González & C.S. Prince",
			Genres:        []string{"Suspenso", "Drama", "Basada en hechos reales"},
			Duration:      "Miniserie (6 episodios)",
			Country:       "Colombia",
		},
		{
			ID:            "cat-book-1",
			Category:      domain.CategoryBook,
			Title:         "El olvido que seremos",
			Year:          2006,
			Description:   "Conmovedor relato biográfico en el que Héctor Abad Faciolince reconstruye la vida de su padre, el médico y defensor de los derechos humanos Héctor Abad Gómez, en la Medellín de los años setenta y ochenta.",
			PosterURL:     "https://images.unsplash.com/photo-1544947950-fa07a98d237f?auto=format&fit=crop&w=600&q=80",
			AverageRating: 4.9,
			RatingsCount:  730,
			Cast:          []string{"Héctor Abad Gómez", "Héctor Abad Faciolince", "Cecilia Faciolince", "Marta Abad"},
			Producer:      "Editorial Planeta / Alfaguara",
			Creator:       "Héctor Abad Faciolince",
			Genres:        []string{"Biografía", "Literatura colombiana", "Memoria"},
			Duration:      "274 páginas",
			Country:       "Colombia",
		},
		{
			ID:            "cat-book-2",
			Category:      domain.CategoryBook,
			Title:         "Delirio",
			Year:          2004,
			Description:   "Aguilar regresa a casa después de un viaje corto de trabajo y encuentra que su esposa Agustina ha perdido la razón. Para ayudarla a salir de la crisis, intenta descubrir qué ocurrió durante esos días, entrelazando cuatro voces en la Bogotá de finales del siglo XX.",
			PosterURL:     "https://images.unsplash.com/photo-1512820790803-83ca734da794?auto=format&fit=crop&w=600&q=80",
			AverageRating: 4.6,
			RatingsCount:  310,
			Cast:          []string{"Agustina Londoño", "Aguilar", "Midas McAlister", "Nicolás Portulinus"},
			Producer:      "Editorial Alfaguara (Premio Alfaguara de Novela)",
			Creator:       "Laura Restrepo",
			Genres:        []string{"Novela", "Ficción contemporánea", "Suspenso psicológico"},
			Duration:      "344 páginas",
			Country:       "Colombia",
		},
	}
}
