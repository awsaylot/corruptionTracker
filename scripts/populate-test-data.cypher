// Clear existing data
MATCH (n) DETACH DELETE n;

// Create Person nodes with corruption scores and other properties
CREATE (trump:Person {
    name: "Donald Trump",
    title: "Former US President",
    corruption_score: 8.5,
    net_worth: "2.5B",
    nationality: "American",
    dob: "1946-06-14",
    aliases: ["The Donald", "45"],
    notes: "Former US President and business mogul"
})

CREATE (putin:Person {
    name: "Vladimir Putin",
    title: "Russian President",
    corruption_score: 9.2,
    net_worth: "70B",
    nationality: "Russian",
    dob: "1952-10-07",
    notes: "Long-serving Russian leader"
})

CREATE (gates:Person {
    name: "Bill Gates",
    title: "Microsoft Co-founder",
    corruption_score: 3.2,
    net_worth: "105B",
    nationality: "American",
    dob: "1955-10-28",
    notes: "Tech pioneer and philanthropist"
})

CREATE (epstein:Person {
    name: "Jeffrey Epstein",
    title: "Financial Manager",
    corruption_score: 9.8,
    net_worth: "500M",
    nationality: "American",
    dob: "1953-01-20",
    deceased: true,
    notes: "Convicted sex offender with high-profile connections"
})

CREATE (maxwell:Person {
    name: "Ghislaine Maxwell",
    title: "Socialite",
    corruption_score: 9.5,
    nationality: "British",
    dob: "1961-12-25",
    notes: "Convicted of sex trafficking"
})

CREATE (musk:Person {
    name: "Elon Musk",
    title: "CEO of Tesla & SpaceX",
    corruption_score: 5.5,
    net_worth: "234B",
    nationality: "American",
    dob: "1971-06-28",
    notes: "Tech entrepreneur and billionaire"
})

// Create Organization nodes
CREATE (microsoft:Organization {
    name: "Microsoft",
    type: "Technology Company",
    founded: "1975",
    hq: "Redmond, WA",
    revenue: "168B",
    corruption_score: 2.8
})

CREATE (tesla:Organization {
    name: "Tesla",
    type: "Automotive & Energy Company",
    founded: "2003",
    hq: "Austin, TX",
    revenue: "81.5B",
    corruption_score: 4.2
})

CREATE (trump_org:Organization {
    name: "The Trump Organization",
    type: "Real Estate Company",
    founded: "1923",
    hq: "New York, NY",
    corruption_score: 7.8
})

// Create relationships
CREATE (trump)-[:FOUNDED {year: 1971}]->(trump_org)
CREATE (gates)-[:FOUNDED {year: 1975}]->(microsoft)
CREATE (musk)-[:LEADS {since: 2008}]->(tesla)

// Business relationships
CREATE (trump)-[:BUSINESS_DEAL {
    type: "Real Estate",
    year: 2008,
    value: "95M",
    notes: "Palm Beach mansion sale"
}]->(putin)

CREATE (gates)-[:MET_WITH {
    frequency: "Multiple times",
    years: "2011-2014",
    context: "Philanthropy discussions"
}]->(epstein)

CREATE (trump)-[:ASSOCIATED_WITH {
    years: "1987-2019",
    context: "Social circles",
    locations: ["Mar-a-Lago", "New York"]
}]->(epstein)

CREATE (maxwell)-[:COLLABORATED_WITH {
    years: "1992-2019",
    context: "Close associates"
}]->(epstein)

CREATE (musk)-[:BUSINESS_DEAL {
    type: "Technology",
    year: 2022,
    value: "44B",
    context: "Twitter acquisition"
}]->(gates)

// Add some media organizations for context
CREATE (fox:Organization {
    name: "Fox News",
    type: "Media Company",
    founded: "1996",
    corruption_score: 6.5
})

CREATE (cnn:Organization {
    name: "CNN",
    type: "Media Company",
    founded: "1980",
    corruption_score: 5.8
})

// Add media relationships
CREATE (trump)-[:APPEARED_ON {
    frequency: "Regular",
    years: "2015-2021",
    sentiment: "Positive"
}]->(fox)

CREATE (trump)-[:APPEARED_ON {
    frequency: "Regular",
    years: "2015-2021",
    sentiment: "Negative"
}]->(cnn)

// Add some financial institutions
CREATE (deutsche:Organization {
    name: "Deutsche Bank",
    type: "Financial Institution",
    founded: "1870",
    corruption_score: 7.2
})

CREATE (trump)-[:BORROWED_FROM {
    years: "2012-2022",
    amount: "2B",
    notes: "Multiple real estate loans"
}]->(deutsche);

// Update timestamps
CALL {
    MATCH (n)
    WHERE n.created_at IS NULL
    SET n.created_at = datetime(),
        n.updated_at = datetime()
};
