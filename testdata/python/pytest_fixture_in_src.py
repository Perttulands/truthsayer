# Production code with pytest fixtures
import pytest

@pytest.fixture
def database():
    return connect_db()

@pytest.fixture(scope="session")
def app():
    return create_app()

def real_function():
    return 42
