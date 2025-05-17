package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"igdb-database/config"
	"igdb-database/db"
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
	baseUrl, err := url.Parse(config.C().ExternalUrl)
	if err != nil {
		log.Fatalf("failed to parse url: %v", err)
	}

	http.HandleFunc(webhook(client.AgeRatingCategories, client))
	http.HandleFunc(webhook(client.AgeRatingContentDescriptions, client))
	http.HandleFunc(webhook(client.AgeRatingContentDescriptionsV2, client))
	http.HandleFunc(webhook(client.AgeRatingOrganizations, client))
	http.HandleFunc(webhook(client.AgeRatings, client))
	http.HandleFunc(webhook(client.AlternativeNames, client))
	http.HandleFunc(webhook(client.Artworks, client))
	http.HandleFunc(webhook(client.CharacterGenders, client))
	http.HandleFunc(webhook(client.CharacterMugShots, client))
	http.HandleFunc(webhook(client.Characters, client))
	http.HandleFunc(webhook(client.CharacterSpecies, client))
	http.HandleFunc(webhook(client.CollectionMemberships, client))
	http.HandleFunc(webhook(client.CollectionMembershipTypes, client))
	http.HandleFunc(webhook(client.CollectionRelations, client))
	http.HandleFunc(webhook(client.CollectionRelationTypes, client))
	http.HandleFunc(webhook(client.Collections, client))
	http.HandleFunc(webhook(client.CollectionTypes, client))
	http.HandleFunc(webhook(client.Companies, client))
	http.HandleFunc(webhook(client.CompanyLogos, client))
	http.HandleFunc(webhook(client.CompanyStatuses, client))
	http.HandleFunc(webhook(client.CompanyWebsites, client))
	http.HandleFunc(webhook(client.Covers, client))
	http.HandleFunc(webhook(client.DateFormats, client))
	http.HandleFunc(webhook(client.EventLogos, client))
	http.HandleFunc(webhook(client.EventNetworks, client))
	http.HandleFunc(webhook(client.Events, client))
	http.HandleFunc(webhook(client.ExternalGames, client))
	http.HandleFunc(webhook(client.ExternalGameSources, client))
	http.HandleFunc(webhook(client.Franchises, client))
	http.HandleFunc(webhook(client.GameEngineLogos, client))
	http.HandleFunc(webhook(client.GameEngines, client))
	http.HandleFunc(webhook(client.GameLocalizations, client))
	http.HandleFunc(webhook(client.GameModes, client))
	http.HandleFunc(webhook(client.GameReleaseFormats, client))
	http.HandleFunc(webhook(client.Games, client))
	http.HandleFunc(webhook(client.GameStatuses, client))
	http.HandleFunc(webhook(client.GameTimeToBeats, client))
	http.HandleFunc(webhook(client.GameTypes, client))
	http.HandleFunc(webhook(client.GameVersionFeatures, client))
	http.HandleFunc(webhook(client.GameVersionFeatureValues, client))
	http.HandleFunc(webhook(client.GameVersions, client))
	http.HandleFunc(webhook(client.GameVideos, client))
	http.HandleFunc(webhook(client.Genres, client))
	http.HandleFunc(webhook(client.InvolvedCompanies, client))
	http.HandleFunc(webhook(client.Keywords, client))
	http.HandleFunc(webhook(client.Languages, client))
	http.HandleFunc(webhook(client.LanguageSupports, client))
	http.HandleFunc(webhook(client.LanguageSupportTypes, client))
	http.HandleFunc(webhook(client.MultiplayerModes, client))
	http.HandleFunc(webhook(client.NetworkTypes, client))
	http.HandleFunc(webhook(client.PlatformFamilies, client))
	http.HandleFunc(webhook(client.PlatformLogos, client))
	http.HandleFunc(webhook(client.Platforms, client))
	http.HandleFunc(webhook(client.PlatformTypes, client))
	http.HandleFunc(webhook(client.PlatformVersionCompanies, client))
	http.HandleFunc(webhook(client.PlatformVersionReleaseDates, client))
	http.HandleFunc(webhook(client.PlatformVersions, client))
	http.HandleFunc(webhook(client.PlatformWebsites, client))
	http.HandleFunc(webhook(client.PlayerPerspectives, client))
	http.HandleFunc(webhook(client.PopularityTypes, client))
	http.HandleFunc(webhook(client.Regions, client))
	http.HandleFunc(webhook(client.ReleaseDateRegions, client))
	http.HandleFunc(webhook(client.ReleaseDates, client))
	http.HandleFunc(webhook(client.ReleaseDateStatuses, client))
	http.HandleFunc(webhook(client.Screenshots, client))
	http.HandleFunc(webhook(client.Themes, client))
	http.HandleFunc(webhook(client.Websites, client))
	http.HandleFunc(webhook(client.WebsiteTypes, client))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("Hello World!")); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	serverStart := make(chan bool)
	go func() {
		defer close(serverStart)
		log.Printf("starting webhook server on %s", config.C().Address)
		err = http.ListenAndServe(config.C().Address, nil)
		if err != nil {
			log.Fatalf("failed to start webhook server: %v", err)
		}
	}()

	enabledEndpoint := endpoint.AllNames
	enabledEndpoint = slices.DeleteFunc(enabledEndpoint, func(e endpoint.Name) bool {
		return e == endpoint.EPWebhooks || e == endpoint.EPSearch || e == endpoint.EPPopularityPrimitives
	})

	ip := net.ParseIP(baseUrl.Hostname())
	if baseUrl.Hostname() == "localhost" || (ip != nil && ip.IsLoopback()) {
		log.Printf("extral url is localhost. webhook will not be registered")
	} else {
		for _, ep := range enabledEndpoint {
			Url := baseUrl.JoinPath(fmt.Sprintf("/webhook/%s", string(ep)))
			log.Printf("registering webhook \"%s\" to \"%s\"", ep, Url.String())
			_, err = client.Webhooks.Register(ep, config.C().WebhookSecret, Url.String(), endpoint.WebhookMethodCreate)
			if err != nil {
				log.Fatalf("failed to register webhook \"%s\": %v", ep, err)
			}
			_, err = client.Webhooks.Register(ep, config.C().WebhookSecret, Url.String(), endpoint.WebhookMethodUpdate)
			if err != nil {
				log.Fatalf("failed to register webhook \"%s\": %v", ep, err)
			}
			log.Printf("webhook \"%s\" registered", ep)
		}
		log.Printf("all webhook registered")
	}

	<-serverStart
}

func webhook[T any](
	e endpoint.EntityEndpoint[T],
	client *igdb.Client,
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

		if _, ok := any(e).(*endpoint.Games); ok {
			game := any(item).(*pb.Game)
			g, err := db.ConvertGame(game, client)
			if err != nil {
				log.Printf("failed to convert game: %v", err)
			} else {
				_ = db.SaveGame(g)
				log.Printf("game %d aggregated", data.ID)
			}
		}

		err = db.SaveItem(e.GetEndpointName(), item)
		if err != nil {
			log.Printf("failed to save %s: %v", e.GetEndpointName(), err)
			return
		}

		// update associated game
		type gameGetter interface {
			GetGame() *pb.Game
		}

		if v, ok := any(item).(gameGetter); ok {
			game, err := db.GetItemById[pb.Game](endpoint.EPGames, v.GetGame().Id)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				log.Printf("failed to get game: %v", err)
				goto END
			}
			g, err := db.ConvertGame(game, client)
			if err != nil {
				log.Printf("failed to convert game: %v", err)
				goto END
			}
			err = db.SaveGame(g)
			if err != nil {
				log.Printf("failed to save game: %v", err)
				goto END
			}
			log.Printf("game %d aggregated", data.ID)
		}

	END:
		log.Printf("%s %d saved", e.GetEndpointName(), data.ID)
	}
}
