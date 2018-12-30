# Rezerwacje DUW


[![Build Status](https://cloud.drone.io/api/badges/dyrkin/rezerwacje-duw-go/status.svg)](https://cloud.drone.io/dyrkin/rezerwacje-duw-go)
 [![Github Release](https://img.shields.io/badge/release-1.0.6-blue.svg)](https://github.com/dyrkin/rezerwacje-duw-go/releases/tag/v1.0.6)

## Overview

This application will help you to make:
1. Reservation of a visit for making a legalization of foreigners
2. Reservation of a visit to head of department

## Examples
1. Reservation of a visit for making a legalization of foreigners in any city

    ```bash
    $ ./rezerwacje-duw-go-osx application
    ```
    
2. Reservation of a visit for making a legalization of foreigners in cities Wrocław and Legnica

    ```bash
    $ ./rezerwacje-duw-go-osx application city WRO LG
    ```

3. Reservation of a visit to head of LP1 department

    ```bash
    $ ./rezerwacje-duw-go-osx headof department LP1
    ```

4. Reservation of a visit to head of LP2 department

    ```bash
    $ ./rezerwacje-duw-go-osx headof department LP2
    ```

## To begin reservation

1. download binary file from the [releases](https://github.com/dyrkin/rezerwacje-duw-go/releases) page.

2. download [user.yml.template](https://raw.githubusercontent.com/dyrkin/rezerwacje-duw-go/v1.0.6/user.yml.template) and [application.yml](https://raw.githubusercontent.com/dyrkin/rezerwacje-duw-go/v1.0.6/application.yml) and place them in the same folder where the binary was downloaded.

3. rename `user.yml.template` to `user.yml`.

4. edit `user.yml` file providing your credentials and information required for reservation.
5. run the application
  
    1. for **windows**: 
         
        ```$ rezerwacje-duw-go-win{32/64}.exe```

    2. for **os x**:

        ```$ ./rezerwacje-duw-go-osx```

    3. for **linux**:

        ```$ ./rezerwacje-duw-go-linux{32/64}```

    with the following options:

6. wait until the message appears `Reservation completed for Wrocław, slot 123456 and time 2018-11-03 11:15:00. Check your email or DUW site`.

7. close the app by pressing any key.
8. check your email or DUW site.

## To run application from the source code

1. install `golang` of min version `1.11`.

    ```$ brew install go```

2. go to the application folder.
3. copy `user.yml.template`

   ```$ cp user.yml.template user.yml``` 

4. edit `user.yml` file providing your credentials and information required for reservation.
5. run the application

   ```$ go run main.go```