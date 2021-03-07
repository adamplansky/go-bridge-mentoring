https://cloud.google.com/pubsub/docs/emulator
```shell
gcloud components install pubsub-emulator
gcloud components update
gcloud beta emulators pubsub start --project=PUBSUB_PROJECT_ID


gcloud beta emulators pubsub start --project=gdrive-adam-plansky
gcloud beta emulators pubsub env-init

# set env 
PUBSUB_EMULATOR_HOST=::1:8960
```