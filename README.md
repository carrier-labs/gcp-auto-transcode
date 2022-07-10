# gcp-auto-transcode
Automatic processes to watch and transcode video files in a Storage Bucket into standard optimised formats

## Storage Function

- Type: Storage Trigger
- Event: google.storage.object.finalize
- Entry Point: WatchStorageBucket
- Watches:
  - `media/upload/*.*`
  - `media/video/*/og-*.*`
    - Triggers PubSub-> `transcode-queue`
  - `media/video/*/og-*.*`

## Transcode Job Queue

- Type: PubSub
- Entrypoint: PubSubTranscodeQueue
- Topic: `transcode-queue`
  - Triggers: TranscoderAPI
    - Returns: PubSub-> `transcode-complete`

## Transcode Job Result

- Type: PubSub
- Entrypoint: PubSubTranscodeResult
- Topic: `transcode-result`
