from rest_framework import serializers
from dateutil import parser as date_parser


class MovieSerializerFromJson(serializers.Serializer):
    """
    This serializer is designed to take the output from a JSON file of movies and
    translate it into a Movie object
    """

    id = serializers.CharField()
    title = serializers.CharField()
    year = serializers.SerializerMethodField()
    poster_url = serializers.CharField(source="poster_path")
    description = serializers.CharField(source="overview")

    def get_year(self, obj):
        date_obj = date_parser.parse(obj["release_date"])

        return date_obj.year
