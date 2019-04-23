#HowTo

1. clone the project
2. copy .env_example to .env
3. fill the .env 
   1. LISTEN_IP - can be left 0.0.0.0
   2. LISTEN_PORT - set whatever you want e.g 12345
   3. TRELLO_APIKEY - get it from here https://trello.com/app-key
   4. TRELLO_TOKEN - get it from here https://trello.com/app-key (generate token manually)
   5. TRELLO_USERNAME - your username in trello. the daemon will listen for events under your username and api credentials
   6. TRELLO_WEBHOOK_URL - put here absolute path where daemon is available e.g http://ip_or_domain:12345
   7. SLACK_TOKEN - get it from here https://api.slack.com/custom-integrations/legacy-tokens
   8. TRELLO_CLEAR_PREVIOUS_HOOKS - set to 1 to force daemon to clear all hooks on all available dashboards before setting own
   9. LOGIN_RELATION_1 - relation of trello login to slack id. separator is | . exmaple "@test|U123456". You can create up to 100 such sections. just increment postfix
4. go get  
5. go build
6. start the daemon. it will setup or check hooks on its start

P.S. slack user ids for relations can be found here https://api.slack.com/methods/auth.test/test
