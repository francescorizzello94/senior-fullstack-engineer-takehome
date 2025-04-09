# (Senior) Fullstack Engineer (m/w/d) @OofOne Takehome

As part of our application process, we'd like to see you approach to technical challenges by giving you a small assignment that resembels the challenges awaiting you once you join OofOne. It should take you no more than a few hours to complete the assignment, but any extra polish or features you might want to put in will not go unnoticed.

## [](https://github.com/OofOne-SE/senior-software-engineer-takehome#the-assignment)The assignment

You will find two files in this repository (besides this `README`): `columns.yaml` and `weather.dat`.
`weather.dat` contains mock weather data collected over several months. `columns.yaml` contains the headers for the given data.

Your task is to develop a small API using Go that works with the data as if it was realtime data.
Write a small programm in the scripting language of your choice that iterates over the rows in the data file and sends them to an endpoint of yours one by one.
The endpoint should accept the raw data and store it in a structured manner.

Please also create endpoints to:
- Retrieve the weather data for a given day
- Retrieve the weather data for a range of days
- Expose a websocket connection that transmits the latest data

Write tests as you find necessary and add simple documentation.
Use any database.

Consider the case that the data stream might increase in frequency in the future and the application will need to store larger amounts of data.

## [](https://github.com/OofOne-SE/senior-software-engineer-takehome#requirements)Requirements

You may choose whatever technologies you prefer, the only requirement is Go as the backend language.

If you have any questions, please ask!

To complete your takehome, please fork this repo and commit your work to your fork. When you are ready for us to look at it, give us access to your fork so we can review and run it.
