from neo4j import GraphDatabase


def insert_ATMs():
    query = """
    LOAD CSV WITH HEADERS FROM 'file:///csv/atm.csv' AS row
    MERGE (a:ATM {
        ATM_id: row.ATM_id, 
        loc_latitude: toFloat(row.loc_latitude), 
        loc_longitude: toFloat(row.loc_longitude), 
        city: row.city, 
        country: row.country
    });
    """


def populate_all():
    uri = "bolt://localhost:7687"  # Replace with your Neo4j connection URI
    user = "neo4j"  # Replace with your Neo4j username
    password = "password"  # Replace with your Neo4j password

    driver = GraphDatabase.driver(uri, auth=(user, password))

    with driver.session() as session:
        session.write_transaction(insert_ATMs)

    driver.close()


if __name__ == "__main__":
    populate_all()
