# Production code without pytest fixtures
import os

def get_database():
    return connect_db()

def create_app():
    # Not a pytest.fixture, just a factory
    return App()
