package porygion

import (
	"fmt"
	"image"
	"math/rand"

	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	simplex "github.com/ojrac/opensimplex-go"
)

// RegionMap represents a generated region map.
type RegionMap struct {
	PixelWidth  int
	PixelHeight int
	Elevations  [][]float64
	Cities      []Tile
	Routes      []Tile
}

// GenerateRegionMap generates a new complete region map.
func GenerateRegionMap(seed int64, pixelWidth, pixelHeight int, numCities int) (RegionMap, error) {
	rand.Seed(seed)
	elevations := getNewElevationMap(pixelWidth, pixelHeight)
	generateElevations(elevations)
	validTiles := getValidLandmarkTiles(elevations)
	partitions := partitionTilesByLocation(100, 100, pixelWidth/8, pixelHeight/8, validTiles)
	cities := generateCities(partitions, numCities)
	cityClusters, err := clusterCities(cities)
	if err != nil {
		return RegionMap{}, err
	}
	routes := generateRoutes(cityClusters)
	return RegionMap{
		PixelWidth:  pixelWidth,
		PixelHeight: pixelHeight,
		Elevations:  elevations,
		Cities:      cities,
		Routes:      routes,
	}, nil
}

// GenerateBaseRegionMap generates a new region map containing only elevations.
func GenerateBaseRegionMap(seed int64, pixelWidth, pixelHeight int) RegionMap {
	rand.Seed(seed)
	elevations := getNewElevationMap(pixelWidth, pixelHeight)
	generateElevations(elevations)
	return RegionMap{
		PixelWidth:  pixelWidth,
		PixelHeight: pixelHeight,
		Elevations:  elevations,
	}
}

// GenerateRegionMapWithCities generates a new region map with new city locations, using
// the provided region map.
func GenerateRegionMapWithCities(seed int64, numCities int, regionMap RegionMap) RegionMap {
	rand.Seed(seed)
	validTiles := getValidLandmarkTiles(regionMap.Elevations)
	partitions := partitionTilesByLocation(100, 100, regionMap.PixelWidth/8, regionMap.PixelHeight/8, validTiles)
	cities := generateCities(partitions, numCities)
	regionMap.Cities = cities
	return regionMap
}

// GenerateRegionMapWithRoutes generates a new region map with new route locations, using
// the provided region map.
func GenerateRegionMapWithRoutes(seed int64, regionMap RegionMap) (RegionMap, error) {
	rand.Seed(seed)
	cityClusters, err := clusterCities(regionMap.Cities)
	if err != nil {
		return RegionMap{}, err
	}
	routes := generateRoutes(cityClusters)
	regionMap.Routes = routes
	return regionMap, nil
}

// RenderBaseRegionMap renders a region map using only its elevations.
func RenderBaseRegionMap(regionMap RegionMap) image.Image {
	img := renderRegionMapImage(regionMap.Elevations, []Tile{}, []Tile{})
	return img
}

// RenderRegionMapWithCities renders a region map using only its elevations and cities.
func RenderRegionMapWithCities(regionMap RegionMap) image.Image {
	img := renderRegionMapImage(regionMap.Elevations, regionMap.Cities, []Tile{})
	return img
}

// RenderFullRegionMap renders a full region map.
func RenderFullRegionMap(regionMap RegionMap) image.Image {
	img := renderRegionMapImage(regionMap.Elevations, regionMap.Cities, regionMap.Routes)
	return img
}

func getNewElevationMap(width, height int) [][]float64 {
	elevations := make([][]float64, width)
	for i := range elevations {
		elevations[i] = make([]float64, height)
	}
	return elevations
}

func generateElevations(elevations [][]float64) {
	baseNoise := simplex.New(rand.Int63())
	secondaryNoise := simplex.New(rand.Int63())
	jitterNoise := simplex.New(rand.Int63())
	jitterCoeffNoise := simplex.New(rand.Int63())
	for i := range elevations {
		for j := range elevations[i] {
			baseElevation := baseNoise.Eval2(float64(i)/100.0, float64(j)/100.0) + 0.2
			secondaryElevation := secondaryNoise.Eval2(float64(i)/20.0, float64(j)/20.0) * 0.15
			jitterElevation := jitterNoise.Eval2(float64(i)/15.0, float64(j)/15.0)
			jitterCoeff := jitterCoeffNoise.Eval2(float64(i)/50.0, float64(j)/50.0) * 0.6
			elevation := baseElevation + secondaryElevation + jitterElevation*jitterCoeff
			elevations[i][j] = elevation
		}
	}
}

func getValidLandmarkTiles(elevations [][]float64) []Tile {
	validTiles := []Tile{}
	tilesWidth := len(elevations) / 8
	tilesHeight := len(elevations[0]) / 8
	for i := 0; i < tilesWidth; i++ {
		for j := 0; j < tilesHeight; j++ {
			// A tile is valid if it has at least a certain number
			// of non-water pixels.
			numLandPixels := 0
			found := false
			for x := 0; x < 8; x++ {
				for y := 0; y < 8; y++ {
					px := i*8 + x
					py := j*8 + y
					if elevations[px][py] >= 0 {
						numLandPixels++
						if numLandPixels > 20 {
							validTiles = append(validTiles, Tile{i, j})
							found = true
							break
						}
					}
				}
				if found {
					break
				}
			}
		}
	}
	return validTiles
}

func partitionTilesByLocation(partitionWidth, partitionHeight, tileWidth, tileHeight int, tiles []Tile) map[string][]Tile {
	// Groups tiles into separate partitions, based on a grid.
	partitions := map[string][]Tile{}
	for _, t := range tiles {
		partitionX := t.X / partitionWidth
		partitionY := t.Y / partitionHeight
		key := fmt.Sprintf("%d:%d", partitionX, partitionY)
		if _, ok := partitions[key]; !ok {
			partitions[key] = []Tile{}
		}
		partitions[key] = append(partitions[key], t)
	}
	return partitions
}

func generateCities(partitions map[string][]Tile, numCities int) []Tile {
	// First, get a randomized order of the partitions.
	partitionKeys := make([]string, len(partitions))
	i := 0
	for k := range partitions {
		partitionKeys[i] = k
		i++
	}
	rand.Shuffle(len(partitionKeys), func(i, j int) { partitionKeys[i], partitionKeys[j] = partitionKeys[j], partitionKeys[i] })

	// Loop through partitions, placing one city at a time.
	cities := map[Tile]bool{}
	for c := 0; c < numCities; c++ {
		partition := partitions[partitionKeys[c%len(partitionKeys)]]
		// Attempt to place the city many times, in case several attempts fail,
		// due to contraints.
		for i := 0; i < 50; i++ {
			if city, ok := tryPickCityTile(partition); ok {
				if _, ok = cities[city]; !ok {
					cities[city] = true
					break
				}
			}
		}
	}
	result := make([]Tile, len(cities))
	i = 0
	for city := range cities {
		result[i] = city
		i++
	}
	return result
}

func tryPickCityTile(partition []Tile) (Tile, bool) {
	// Pick a random tile from the partition, and evaluate whether or not
	// we can place a city there.
	for j := 0; j < 50; j++ {
		candidate := partition[rand.Intn(len(partition))]
		// Only allow cities on a 2x2 grid, to avoid adjacent cities
		// and routes.
		if candidate.X%2 != 1 || candidate.Y%2 != 1 {
			continue
		}
		// Don't allow cities to be placed where the in-game UI elements allow.
		if candidate.X < 1 || candidate.Y < 2 || candidate.X > 28 || candidate.Y > 16 {
			continue
		}
		if candidate.X > 14 && candidate.Y > 14 {
			continue
		}
		if candidate.X > 19 && candidate.Y < 5 {
			continue
		}
		return candidate, true
	}
	return Tile{}, false
}

func clusterCities(cities []Tile) ([][]Tile, error) {
	// Cluster the cities into 2 groups, using k-means.
	var points clusters.Observations
	for _, city := range cities {
		points = append(points, clusters.Coordinates{float64(city.X), float64(city.Y)})
	}
	km := kmeans.New()
	clusters, err := km.Partition(points, 2)
	if err != nil {
		return [][]Tile{}, fmt.Errorf("Failed to cluster cities: %s", err)
	}
	cityClusters := make([][]Tile, len(clusters))
	for i, cluster := range clusters {
		for _, o := range cluster.Observations {
			coords := o.Coordinates()
			cityClusters[i] = append(cityClusters[i], Tile{int(coords[0]), int(coords[1])})
		}
	}
	return cityClusters, nil
}

func generateRoutes(cityClusters [][]Tile) []Tile {
	routeTiles := map[Tile]bool{}
	// Connect cities within each cluster to each other.
	for _, cities := range cityClusters {
		if len(cities) < 2 {
			continue
		}
		connectedCities := map[Tile]bool{}
		var city *Tile
		var firstCity *Tile
		var lastCity Tile
		for len(connectedCities) < len(cities) {
			if city == nil {
				// Get an unconnected city.
				for _, c := range cities {
					if _, ok := connectedCities[c]; !ok {
						city = &Tile{c.X, c.Y}
						break
					}
				}
				firstCity = &Tile{city.X, city.Y}
			}
			lastCity = *city
			// Find the nearest unconnected city, and connect it.
			var nearestCity *Tile
			minDistance := 99999
			for _, other := range cities {
				if other == *city {
					continue
				}
				if _, ok := connectedCities[other]; ok {
					continue
				}
				dist := city.Distance(other)
				if dist < minDistance {
					minDistance = dist
					nearestCity = &Tile{other.X, other.Y}
				}
			}
			if nearestCity == nil {
				continue
			}
			connectCities(*city, *nearestCity, routeTiles)
			connectedCities[*city] = true
			connectedCities[*nearestCity] = true
			*city = *nearestCity
		}
		connectCities(*firstCity, lastCity, routeTiles)
	}

	// Connect the two clusters of cities together by
	// connecting the two nearest cities.
	minDistance := 99999
	var cityA Tile
	var cityB Tile
	for _, a := range cityClusters[0] {
		for _, b := range cityClusters[1] {
			dist := a.Distance(b)
			if dist < minDistance {
				minDistance = dist
				cityA = a
				cityB = b
			}
		}
	}
	connectCities(cityA, cityB, routeTiles)

	// Return a slice of tiles, rather than a map.
	result := make([]Tile, len(routeTiles))
	i := 0
	for tile := range routeTiles {
		result[i] = tile
		i++
	}
	return result
}

func connectCities(cityA Tile, cityB Tile, routeTiles map[Tile]bool) {
	if rand.Intn(2) == 0 {
		start := connectHorizontalRoute(cityA, cityB, routeTiles)
		connectVerticalRoute(start, cityB, routeTiles)
	} else {
		start := connectVerticalRoute(cityA, cityB, routeTiles)
		connectHorizontalRoute(start, cityB, routeTiles)
	}
}

func connectHorizontalRoute(start Tile, end Tile, routeTiles map[Tile]bool) Tile {
	inc := 1
	if start.X > end.X {
		inc = -1
	}
	for i := start.X; i != end.X; i += inc {
		t := Tile{i, start.Y}
		routeTiles[t] = true
	}
	return Tile{end.X, start.Y}
}

func connectVerticalRoute(start Tile, end Tile, routeTiles map[Tile]bool) Tile {
	inc := 1
	if start.Y > end.Y {
		inc = -1
	}
	for j := start.Y; j != end.Y; j += inc {
		t := Tile{start.X, j}
		routeTiles[t] = true
	}
	return Tile{start.X, end.Y}
}
