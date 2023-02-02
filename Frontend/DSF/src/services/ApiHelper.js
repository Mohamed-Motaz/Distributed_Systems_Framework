import axios from 'axios';


export default class ApiHelper {


    static async func(mediaType) {
        let response;

        await axios.get(`https://api.themoviedb.org/3/trending/${mediaType}/week?api_key=${API_KEY}`)
            .then((value) => response = value)
            .catch((error) => console.log("ERROR: " + error));

        return response;

    }




}