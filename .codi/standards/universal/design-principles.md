---
language: universal
category: universal
applies_to: []
---

# Design Principles

Based on [Software Engineering Principles](https://www.callicoder.com/software-development-principles/), [SOLID Principles](https://designgurus.io/blog/essential-software-design-principles-you-should-know-before-the-interview), and [System Design Principles](https://blog.paulserban.eu/post/system-design-principles-dry-solid-yagni-and-design-patterns/).

## Quick Reference

- **DRY** - Don't Repeat Yourself
- **KISS** - Keep It Simple, Stupid
- **YAGNI** - You Aren't Gonna Need It
- **SOLID** - Five principles for object-oriented design
- **Separation of Concerns** - Each module has one responsibility

## Core Principles

### 1. DRY (Don't Repeat Yourself)

```
# ❌ Duplication
def calculate_user_discount(user):
    if user.type == "premium":
        return 0.2
    return 0.1

def calculate_order_discount(order):
    if order.user.type == "premium":
        return 0.2
    return 0.1

# ✅ Single source of truth
def get_discount_rate(user):
    return 0.2 if user.type == "premium" else 0.1

def calculate_user_discount(user):
    return get_discount_rate(user)

def calculate_order_discount(order):
    return get_discount_rate(order.user)
```

**When to apply**:
- ✅ Business logic duplication
- ✅ Data transformation patterns
- ❌ Coincidental duplication (looks same, different meaning)

### 2. KISS (Keep It Simple, Stupid)

```
# ❌ Over-engineered
class UserFactory:
    def __init__(self, builder, validator, mapper):
        self.builder = builder
        self.validator = validator
        self.mapper = mapper
    
    def create(self, data):
        validated = self.validator.validate(data)
        mapped = self.mapper.map(validated)
        return self.builder.build(mapped)

# ✅ Simple and clear
def create_user(name, email):
    if not email or "@" not in email:
        raise ValueError("Invalid email")
    return User(name=name, email=email)
```

**Guidelines**:
- Prefer simple solutions over clever ones
- Optimize for readability, not cleverness
- Three similar lines > premature abstraction

### 3. YAGNI (You Aren't Gonna Need It)

```
# ❌ Building for future (that never comes)
class User:
    def __init__(self):
        self.name = None
        self.email = None
        self.phone = None  # Not needed yet
        self.address = None  # Not needed yet
        self.preferences = {}  # Not needed yet

# ✅ Build what you need now
class User:
    def __init__(self, name, email):
        self.name = name
        self.email = email
# Add fields when actually needed
```

**When to apply**:
- Don't add features "just in case"
- Don't build abstractions before second use
- Don't design for hypothetical requirements

### 4. SOLID Principles

#### S - Single Responsibility Principle

```
# ❌ Multiple responsibilities
class User:
    def save(self):
        # Database logic
        pass
    
    def send_email(self):
        # Email logic
        pass
    
    def generate_report(self):
        # Report logic
        pass

# ✅ Single responsibility
class User:
    def __init__(self, name, email):
        self.name = name
        self.email = email

class UserRepository:
    def save(self, user):
        # Database logic
        pass

class EmailService:
    def send_welcome_email(self, user):
        # Email logic
        pass
```

#### O - Open/Closed Principle

```
# ❌ Modify existing code for new types
def calculate_area(shape):
    if shape.type == "circle":
        return 3.14 * shape.radius ** 2
    elif shape.type == "rectangle":
        return shape.width * shape.height
    # Add new elif for each shape

# ✅ Extend without modifying
class Shape:
    def area(self):
        raise NotImplementedError

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius
    
    def area(self):
        return 3.14 * self.radius ** 2

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height
    
    def area(self):
        return self.width * self.height
```

#### L - Liskov Substitution Principle

```
# ❌ Violates LSP
class Bird:
    def fly(self):
        pass

class Penguin(Bird):
    def fly(self):
        raise Exception("Penguins can't fly")  # Breaks contract

# ✅ Follows LSP
class Bird:
    def move(self):
        pass

class FlyingBird(Bird):
    def move(self):
        self.fly()

class Penguin(Bird):
    def move(self):
        self.swim()
```

#### I - Interface Segregation Principle

```
# ❌ Fat interface
class Worker:
    def work(self):
        pass
    
    def eat(self):
        pass
    
    def sleep(self):
        pass

class Robot(Worker):
    def eat(self):
        raise NotImplementedError  # Robot doesn't eat

# ✅ Segregated interfaces
class Workable:
    def work(self):
        pass

class Eatable:
    def eat(self):
        pass

class Human(Workable, Eatable):
    def work(self):
        pass
    
    def eat(self):
        pass

class Robot(Workable):
    def work(self):
        pass
```

#### D - Dependency Inversion Principle

```
# ❌ Depend on concrete implementation
class UserService:
    def __init__(self):
        self.db = PostgresDatabase()  # Tight coupling
    
    def get_user(self, id):
        return self.db.query(f"SELECT * FROM users WHERE id = {id}")

# ✅ Depend on abstraction
class Database:
    def query(self, sql):
        raise NotImplementedError

class PostgresDatabase(Database):
    def query(self, sql):
        # Postgres implementation
        pass

class UserService:
    def __init__(self, database: Database):
        self.db = database  # Depends on interface
    
    def get_user(self, id):
        return self.db.query(f"SELECT * FROM users WHERE id = {id}")
```

### 5. Separation of Concerns

```
# ❌ Mixed concerns
def process_order(order_data):
    # Validation
    if not order_data.get("user_id"):
        return {"error": "Missing user_id"}
    
    # Business logic
    total = sum(item["price"] * item["quantity"] for item in order_data["items"])
    
    # Database
    db.execute("INSERT INTO orders ...")
    
    # Email
    send_email(order_data["user_id"], "Order confirmed")
    
    return {"success": True}

# ✅ Separated concerns
def validate_order(order_data):
    if not order_data.get("user_id"):
        raise ValueError("Missing user_id")

def calculate_total(items):
    return sum(item["price"] * item["quantity"] for item in items)

def save_order(order):
    return repository.save(order)

def send_order_confirmation(user_id):
    email_service.send(user_id, "Order confirmed")

def process_order(order_data):
    validate_order(order_data)
    total = calculate_total(order_data["items"])
    order = save_order({"total": total, ...})
    send_order_confirmation(order_data["user_id"])
    return order
```

**Layers**:
- Presentation (API, UI)
- Business Logic (domain rules)
- Data Access (database, external APIs)

## Common Mistakes

### Mistake 1: Premature Abstraction

```
# ❌ Abstraction before pattern emerges
def process_a():
    return base_process("A")

def process_b():
    return base_process("B")

def base_process(type):
    # Complex generic logic
    pass

# ✅ Wait for third use
def process_a():
    # Implementation A
    pass

def process_b():
    # Implementation B
    pass

# Extract common logic after pattern is clear
```

### Mistake 2: Over-applying DRY

```
# ❌ DRY gone wrong (coincidental duplication)
def format_user_name(user):
    return user.name.strip().title()

def format_product_name(product):
    return product.name.strip().title()

# Looks same but different meaning
# Changing user formatting shouldn't affect products

# ✅ Keep separate if they may diverge
def format_user_name(user):
    return user.name.strip().title()

def format_product_name(product):
    return product.name.strip().upper()  # Products: all caps
```

### Mistake 3: Ignoring YAGNI

```
# ❌ Building for unknown future
class ConfigManager:
    def __init__(self):
        self.providers = []  # Support multiple providers
        self.cache = {}      # Caching layer
        self.fallback = {}   # Fallback config
        # Currently only need simple key-value config

# ✅ Start simple
config = {
    "api_key": "...",
    "timeout": 30
}
# Add complexity when actually needed
```

---

**References**:
- [Software Engineering Principles Guide](https://www.callicoder.com/software-development-principles/)
- [SOLID Principles Explained](https://designgurus.io/blog/essential-software-design-principles-you-should-know-before-the-interview)
- [System Design Principles](https://blog.paulserban.eu/post/system-design-principles-dry-solid-yagni-and-design-patterns/)
- [Engineering with SOLID, DRY, KISS, YAGNI](https://dev.to/andrey_s/engineering-with-solid-dry-kiss-yagni-and-grasp-136a)
