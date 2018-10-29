# IMT2681 Cloud Technologies autumn 2018 - Assignment 1
This assignment revolves about creating an API for retrievment and input of IGC track information.

Done by Håkon Schia (haakosc, 480631).


# Usage


```/paragliding/api/```

**GET**: Returns information about the API.


```/paragliding/api/track/```

**POST**: Adds an IGC tracks to the API, given by a valid URL.

**GET**: Returns an array of the IDs currently in the memory of the API.


```/paragliding/api/track/<ID>```

**GET**: Returns information about the IGC track with the given ID (the internal ID used). This is a numeric ID, starting from 1.


```/paragliding/api/track/<ID>/<field>```

**GET**: Returns the relevant field for the given ID. The valid fields are: H_date, pilot, glider, glider_id, track_length.

# Heroku
Deployed on Heroku under the URL: https://warm-taiga-87322.herokuapp.com/

# MongoDB
The driver choice selected is motivated by me not knowing the difference and not having time to learn the difference because there is way too much other stuff to learn all by myself, so I just chose one of them.
