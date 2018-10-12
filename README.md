# IMT2681 Cloud Technologies autumn 2018 - Assignment 1
This assignment revolves about creating an API for retrievment and input of IGC track information.

Done by Håkon Schia (haakosc, 480631).


# Usage


```/igcinfo/api/```

**GET**: Returns information about the API.


```/igcinfo/api/igc/```

**POST**: Adds an IGC tracks to the API, given by a valid URL.

**GET**: Returns an array of the IDs currently in the memory of the API.


```/igcinfo/api/igc/<ID>```

**GET**: Returns information about the IGC track with the given ID (the internal ID used). This is a numeric ID, starting from 1.


```igcinfo/api/igc/<ID>/<field>```

**GET**: Returns the relevant field for the given ID. The valid fields are: H_date, pilot, glider, glider_id, track_length.

# Heroku link
https://warm-taiga-87322.herokuapp.com/
