
# pigeon 

Docker image build :\
```docker build -t pigeon .```

Docker run :\
``` docker run --name pigeon-scheduler -p 4040:4040 --env-file ./sample_env.list pigeon```

### Usage

#### Create Scheduled Task

Example Request

```
{
   "interval":1,
   "intervalType":"DAILY | SECOND",
   "sendAt":"22:19:00",
   "immediately":"true | false",
   "continuous":"true | false",
   "limit":3,
   "execution":{
      "targetUrl":"http://my.custom.link:8080/sub/dir",
      "type":"GET | POST",
      "body":{
         "key1":"value1",
         "key2":"value2"
      },
      "header":{
         "key1":"value1",
         "key2":"value2"
      }
   }
}
```

Example Response

```
{
    "TaskId": "a8194865-bdad-488b-a39b-3df95d68de1d"
}
```

Description : 

| Field               | Description                                                                      | Required |
| ------------------- | -------------------------------------------------------------------------------- | -------- |
| interval            | Time which between two task fire                                                 | *        |
| intervalType        | Which time interval for two task                                                 | *        |
| sendAt              | When will fire the first task                                                    |          |
| immediately         | Fire task when request come to service <br> (**override sendAt for first fire**) |          |
| limit               | How many task will fire                                                          |          |
| continuous          | Fires task every time <br> (**overrides limit**)                                 |          |
| execution           | Target service details                                                           | *        |
| execution.targetUrl | Target endpoint will trigger                                                     | *        |
| execution.type      | Request type                                                                     | *        |
| execution.body      | Request body <br> (**necessary if type is POST**)                                |          |
| execution.header    | Request header                                                                   | *        |
