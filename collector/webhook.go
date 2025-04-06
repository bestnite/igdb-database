package collector

import (
	"encoding/json"
	"fmt"
	"igdb-database/config"
	"igdb-database/db"
	"igdb-database/model"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"slices"

	pb "github.com/bestnite/go-igdb/proto"

	"github.com/bestnite/go-igdb"
	"github.com/bestnite/go-igdb/endpoint"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func StartWebhookServer(client *igdb.Client) {
	baseUrl, err := url.Parse(config.C().ExtralUrl)
	if err != nil {
		log.Fatalf("failed to parse url: %v", err)
	}

	http.HandleFunc(webhook(client.AgeRatingCategories))
	http.HandleFunc(webhook(client.AgeRatingContentDescriptions))
	http.HandleFunc(webhook(client.AgeRatingContentDescriptionsV2))
	http.HandleFunc(webhook(client.AgeRatingOrganizations))
	http.HandleFunc(webhook(client.AgeRatings))
	http.HandleFunc(webhook(client.AlternativeNames))
	http.HandleFunc(webhook(client.Artworks))
	http.HandleFunc(webhook(client.CharacterGenders))
	http.HandleFunc(webhook(client.CharacterMugShots))
	http.HandleFunc(webhook(client.Characters))
	http.HandleFunc(webhook(client.CharacterSpecies))
	http.HandleFunc(webhook(client.CollectionMemberships))
	http.HandleFunc(webhook(client.CollectionMembershipTypes))
	http.HandleFunc(webhook(client.CollectionRelations))
	http.HandleFunc(webhook(client.CollectionRelationTypes))
	http.HandleFunc(webhook(client.Collections))
	http.HandleFunc(webhook(client.CollectionTypes))
	http.HandleFunc(webhook(client.Companies))
	http.HandleFunc(webhook(client.CompanyLogos))
	http.HandleFunc(webhook(client.CompanyStatuses))
	http.HandleFunc(webhook(client.CompanyWebsites))
	http.HandleFunc(webhook(client.Covers))
	http.HandleFunc(webhook(client.DateFormats))
	http.HandleFunc(webhook(client.EventLogos))
	http.HandleFunc(webhook(client.EventNetworks))
	http.HandleFunc(webhook(client.Events))
	http.HandleFunc(webhook(client.ExternalGames))
	http.HandleFunc(webhook(client.ExternalGameSources))
	http.HandleFunc(webhook(client.Franchises))
	http.HandleFunc(webhook(client.GameEngineLogos))
	http.HandleFunc(webhook(client.GameEngines))
	http.HandleFunc(webhook(client.GameLocalizations))
	http.HandleFunc(webhook(client.GameModes))
	http.HandleFunc(webhook(client.GameReleaseFormats))
	http.HandleFunc(webhook(client.Games))
	http.HandleFunc(webhook(client.GameStatuses))
	http.HandleFunc(webhook(client.GameTimeToBeats))
	http.HandleFunc(webhook(client.GameTypes))
	http.HandleFunc(webhook(client.GameVersionFeatures))
	http.HandleFunc(webhook(client.GameVersionFeatureValues))
	http.HandleFunc(webhook(client.GameVersions))
	http.HandleFunc(webhook(client.GameVideos))
	http.HandleFunc(webhook(client.Genres))
	http.HandleFunc(webhook(client.InvolvedCompanies))
	http.HandleFunc(webhook(client.Keywords))
	http.HandleFunc(webhook(client.Languages))
	http.HandleFunc(webhook(client.LanguageSupports))
	http.HandleFunc(webhook(client.LanguageSupportTypes))
	http.HandleFunc(webhook(client.MultiplayerModes))
	http.HandleFunc(webhook(client.NetworkTypes))
	http.HandleFunc(webhook(client.PlatformFamilies))
	http.HandleFunc(webhook(client.PlatformLogos))
	http.HandleFunc(webhook(client.Platforms))
	http.HandleFunc(webhook(client.PlatformTypes))
	http.HandleFunc(webhook(client.PlatformVersionCompanies))
	http.HandleFunc(webhook(client.PlatformVersionReleaseDates))
	http.HandleFunc(webhook(client.PlatformVersions))
	http.HandleFunc(webhook(client.PlatformWebsites))
	http.HandleFunc(webhook(client.PlayerPerspectives))
	http.HandleFunc(webhook(client.PopularityTypes))
	http.HandleFunc(webhook(client.Regions))
	http.HandleFunc(webhook(client.ReleaseDateRegions))
	http.HandleFunc(webhook(client.ReleaseDates))
	http.HandleFunc(webhook(client.ReleaseDateStatuses))
	http.HandleFunc(webhook(client.Screenshots))
	http.HandleFunc(webhook(client.Themes))
	http.HandleFunc(webhook(client.Websites))
	http.HandleFunc(webhook(client.WebsiteTypes))

	enabledEndpoint := endpoint.AllEndpoints
	enabledEndpoint = slices.DeleteFunc(enabledEndpoint, func(e endpoint.EndpointName) bool {
		return e == endpoint.EPWebhooks || e == endpoint.EPSearch || e == endpoint.EPPopularityPrimitives
	})

	ip := net.ParseIP(baseUrl.Hostname())
	if baseUrl.Hostname() == "localhost" || (ip != nil && ip.IsLoopback()) {
		log.Printf("extral url is localhost. webhook will not be registered")
	} else {
		for _, endp := range enabledEndpoint {
			Url := baseUrl.JoinPath(fmt.Sprintf("/webhook/%s", string(endp)))
			log.Printf("registering webhook \"%s\" to \"%s\"", endp, Url.String())
			err = client.Webhooks.Register(endp, config.C().WebhookSecret, Url.String())
			if err != nil {
				log.Fatalf("failed to register webhook \"%s\": %v", endp, err)
			}
			log.Printf("webhook \"%s\" registered", endp)
		}
		log.Printf("all webhook registered")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("Hello World!")); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	log.Printf("starting webhook server on %s", config.C().Address)
	err = http.ListenAndServe(config.C().Address, nil)
	if err != nil {
		log.Fatalf("failed to start webhook server: %v", err)
	}
}

func webhook[T any](
	e endpoint.EntityEndpoint[T],
) (string, func(w http.ResponseWriter, r *http.Request)) {
	return fmt.Sprintf("/webhook/%s", e.GetEndpointName()), func(w http.ResponseWriter, r *http.Request) {
		secret := r.Header.Get("X-Secret")
		if secret != config.C().WebhookSecret {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
		data := struct {
			ID uint64 `json:"id"`
		}{}
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)
			return
		}
		err = json.Unmarshal(jsonBytes, &data)
		if err != nil {
			log.Printf("failed to unmarshal request body: %v", err)
			return
		}
		if data.ID == 0 {
			return
		}
		item, err := e.GetByID(data.ID)
		if err != nil {
			log.Printf("failed to get %s: %v", e.GetEndpointName(), err)
			return
		}
		oldItem, err := db.GetItemByIGDBID[T](e.GetEndpointName(), data.ID)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Printf("failed to get %s: %v", e.GetEndpointName(), err)
			return
		}
		newItem := model.NewItem(item)
		if oldItem != nil {
			newItem.MId = oldItem.MId
		}
		err = db.SaveItem(e.GetEndpointName(), newItem)
		if err != nil {
			log.Printf("failed to save %s: %v", e.GetEndpointName(), err)
			return
		}

		// update associated game
		type gameGetter interface {
			GetGame() *pb.Game
		}

		if v, ok := any(item).(gameGetter); ok {
			game, err := db.GetItemByIGDBID[pb.Game](endpoint.EPGames, v.GetGame().Id)
			if err != nil && err != mongo.ErrNoDocuments {
				log.Printf("failed to get game: %v", err)
				return
			}
			g, err := db.ConvertGame(game.Item)
			if err != nil {
				log.Printf("failed to convert game: %v", err)
				return
			}
			oldGame, err := db.GetGameByIGDBID(game.Item.Id)
			if err != nil && err != mongo.ErrNoDocuments {
				log.Printf("failed to get game: %v", err)
				return
			}
			g.MId = oldGame.MId
			err = db.SaveGame(g)
			if err != nil {
				log.Printf("failed to save game: %v", err)
				return
			}
		}

		log.Printf("%s %d saved", e.GetEndpointName(), data.ID)
	}
}
