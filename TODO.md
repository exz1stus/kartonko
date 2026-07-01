~~Fixed Nginx config~~

FRONTEND:

- ~~Tags input hints~~
- ~~Upload Form to use grid~~
- ~~Upload image progress bar~~
- Url uploading
- Perspective image card over inteface
- Image page
- Dynamic favicon and title
- Reverse image search, find image origins, suggest replacing image with better resolution
- Saucenao
- Rework log loading
- log grouping
- Single upload if hash matches, route to image page
- Profile page
- Single error handling system
- Tags mouse tooltip overlay
- Single filtering system for searching and boards
- Themes
- Bg image
- Image sound
- image surfing page, add image sources
- image UPDATE method
- image editing page
- Photopea edit image integration
- Uploading update webhook
- Fix captcha not reloading on error upload
- Add Last seen on webhooks
- Crawling from selected resources: telegram groups, discord active channels

BACKEND:

- ~~Delete image query~~
- Cover Api with test
- Boards system
- boards references - each image has origin board - add image to board: upload as primary, add reference, add filter query
- DB transactions, prevent upload abort errors
- Api token for frontend and telegram
- Remove limits for moderator users
- From filename to hash identification, optional names

TGBOT:

- Rewrite tgbot to use webhooks
- Save to kartonko on reaction
