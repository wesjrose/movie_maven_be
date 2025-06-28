from django.http import JsonResponse
from django.core.paginator import Paginator


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

        page = int(request.query_params.get("page", 1))

        json_movie_data = None
        with open(
            os.path.join("movies", "data", "movie_data_2025_06_20.json"), "r"
        ) as f:
            json_movie_data = json.load(f)

        paginator = Paginator(json_movie_data, 20)
        max_pages = paginator.num_pages
        page_data = paginator.get_page(int(page))

        next_page = page + 1 if (page + 1) <= max_pages else None

        movie_data = MovieSerializerFromJson(page_data, many=True)
        movie_data = movie_data.data

        return JsonResponse({"movies": movie_data, "next_page": next_page}, status=200)
