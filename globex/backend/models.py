from sqlalchemy import Column, Integer, String, Text, Boolean
from database import Base

class Company(Base):
    __tablename__ = "companies"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True)
    inn = Column(String, unique=True, index=True) # ИНН
    ogrn = Column(String, unique=True, index=True) # ОГРН
    description = Column(Text)
    industry = Column(String)
    geography = Column(String)
    is_verified = Column(Boolean, default=False)
    # Add more fields as per detailed requirements later
