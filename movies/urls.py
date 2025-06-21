from django.urls import path

from movies.views import MoviesView

urlpatterns = [
    path("movies", MoviesView.as_view(), name="movies"),
]
