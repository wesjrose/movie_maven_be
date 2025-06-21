from django.http import JsonResponse

from rest_framework.views import APIView

from movies.serializers import MovieSerializerFromJson

import json, os


class MoviesView(APIView):
    authentication_classes = []
    permission_classes = []

    def get(self, request):
        """
        This endpoint returns a list of movies
        """
        json_movie_data = None
        with open(
            os.path.join("movies", "data", "movie_data_2025_06_20.json"), "r"
        ) as f:
            json_movie_data = json.load(f)

        movie_data = MovieSerializerFromJson(json_movie_data, many=True)
        movie_data = movie_data.data
        return JsonResponse({"movies": movie_data}, status=200)
