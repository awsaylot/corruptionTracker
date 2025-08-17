// Create constraints for unique names
CREATE CONSTRAINT person_name IF NOT EXISTS
FOR (p:Person) REQUIRE p.name IS UNIQUE;

CREATE CONSTRAINT org_name IF NOT EXISTS
FOR (o:Organization) REQUIRE o.name IS UNIQUE;

CREATE CONSTRAINT event_title IF NOT EXISTS
FOR (e:Event) REQUIRE e.title IS UNIQUE;

// Create indexes for faster lookups
CREATE INDEX person_alias IF NOT EXISTS
FOR (p:Person) ON (p.aliases);

CREATE INDEX org_type IF NOT EXISTS
FOR (o:Organization) ON (o.type);

CREATE INDEX event_date IF NOT EXISTS
FOR (e:Event) ON (e.date);
