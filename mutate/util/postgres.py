"""Utilities for PostgreSQL interaction."""
import logging
import os

import psycopg2
from psycopg2 import extras

logging.basicConfig(level=logging.INFO, format='%(asctime)s,%(msecs)03d %(levelname)-8s [%(filename)s:%(lineno)d] %(message)s',
    datefmt='%Y-%m-%d:%H:%M:%S')
logger = logging.getLogger(__name__)

class Postgres:
    """Helper class for database manipulation."""
    def __init__(self, host: str = None, user: str = None,  # noqa
                 password: str = None, dbname: str = None,
                 port: int = None) -> None:
        """Database manger"""
        args = (host, user, password, dbname, port)
        if all(map(lambda x: x is None, args)):
            self.db_cfg = get_postgres_config()
        else:
            self.db_cfg = dict(host=host,
                               user=user,
                               password=password,
                               dbname=dbname,
                               port=port)
        self.conn = None

    def connect(self) -> None:
        """Connect to the database."""
        self.conn = psycopg2.connect(**self.db_cfg)

    def ensure_connected(self) -> None:
        """Connect to the DB if needed."""
        if not self.conn or self.conn.closed:
            self.connect()

    def fetchone(self, statement: str, args=()):
        """Fetch one element from the database."""
        self.ensure_connected()
        res = None

        try:
            with self.conn.cursor() as cur:
                cur.execute(statement, args)
                res = cur.fetchone()
        except:
            self.conn.rollback()
            raise
        return res

    def fetchall(self, statement: str, args=()) -> list:
        """Return all elements returned by the given query."""
        self.ensure_connected()

        try:
            with self.conn.cursor(cursor_factory=extras.RealDictCursor) as cur:
                cur.execute(statement, args)
                res = cur.fetchall()

        except (psycopg2.OperationalError, psycopg2.InterfaceError) as e:
            logger.error(str(e))
            self.conn.rollback()
            raise
        except psycopg2.DatabaseError as e:
            logger.error(str(e))
            self.conn.rollback()
            raise
        except Exception as e:
            logger.error(str(e))
            self.conn.rollback()
            raise

        return res

    def commit(self, statement: str, args=()) -> None:
        """Execute the given query and commit the changes."""
        self.ensure_connected()

        try:
            with self.conn.cursor() as cur:
                cur.execute(statement, args)
                self.conn.commit()
        except:
            self.conn.rollback()
            raise

    def update_statistic(self, addr: str, stype: str):
        if addr is not None:
            query = """
                INSERT INTO public.utm_asset_metrics
                (id, asset_name, metric, amount) VALUES
                (%s, %s, %s, %s) ON CONFLICT (asset_name, metric) DO
                UPDATE SET amount = public.utm_asset_metrics.amount + 1
            """
            self.commit(
                query, (addr + "_" + stype, addr, stype, 1))


def get_postgres_config() -> dict:
    """Return a dictionary with postgres configuration."""
    return dict(host=os.environ.get('DB_HOST', 'localhost'),
                user=os.environ.get('DB_USER'),
                password=os.environ.get('DB_PASSWORD'),
                dbname=os.environ.get('DB_NAME'),
                port=os.environ.get('DB_PORT', 5432))
