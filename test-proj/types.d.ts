interface User {
  website: string,
  company: Company,
  id: number,
  name: string,
  username: string,
  email: string,
  address: Address,
  phone: string,
}

interface Address {
  suite: string,
  city: string,
  zipcode: string,
  geo: Geo,
  street: string,
}

interface Geo {
  lng: string,
  lat: string,
}

interface Company {
  name: string,
  catchPhrase: string,
  bs: string,
}

interface Album {
  userId: number,
  id: number,
  title: string,
}

