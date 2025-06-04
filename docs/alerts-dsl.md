# FlareGo Alerts DSL

This document describes the Domain Specific Language (DSL) used for defining alert rules in FlareGo.

## Overview

The FlareGo alerts DSL is a simple expression language designed for evaluating conditions over runtime metrics. It supports basic arithmetic operations, comparisons, and boolean logic.

## Syntax

### Basic Grammar (EBNF)

```
Expr   = Or ;
Or     = And { "||" And } ;
And    = Cmp { "&&" Cmp } ;
Cmp    = Add { ( '>' | '>=' | '<' | '<=' | '==' | '!=' ) Add } ;
Add    = Mul { ('+'|'-') Mul } ;
Mul    = Unary { ('*'|'/') Unary } ;
Unary  = [ '!' | '-' ] Primary ;
Primary= Number | Ident | '(' Expr ')' ;
```

### Identifiers

- Must start with a letter or underscore
- Can contain letters, numbers, and underscores
- Case-sensitive
- Examples: `heap_bytes`, `blocked_goroutines`, `gc_pause_ns`

### Operators

#### Comparison Operators

- `>` - Greater than
- `>=` - Greater than or equal
- `<` - Less than
- `<=` - Less than or equal
- `==` - Equal to
- `!=` - Not equal to

#### Arithmetic Operators

- `+` - Addition
- `-` - Subtraction
- `*` - Multiplication
- `/` - Division

#### Boolean Operators

- `&&` - Logical AND
- `||` - Logical OR
- `!` - Logical NOT

### Values

- Numbers: Integer or floating-point
- Identifiers: Metric names
- Boolean: Non-zero is true, zero is false

## Examples

### Basic Comparisons

```yaml
# Alert when blocked goroutines exceed 150
blocked_goroutines > 150

# Alert when heap usage exceeds 512MB
heap_bytes > 536870912

# Alert when GC pause time exceeds 100ms
gc_pause_ns > 100000000
```

### Compound Conditions

```yaml
# Alert when both conditions are true
blocked_goroutines > 100 && heap_bytes > 268435456

# Alert when either condition is true
gc_pause_ns > 50000000 || heap_bytes > 1073741824

# Complex condition with parentheses
(blocked_goroutines > 50 && heap_bytes > 268435456) || gc_pause_ns > 100000000
```

### Arithmetic Expressions

```yaml
# Alert when heap usage exceeds 512MB
heap_bytes / 1024 / 1024 > 512

# Alert when average pause time exceeds threshold
total_pause_ns / gc_count > 100000000

# Alert when memory pressure is high
heap_bytes / heap_objects > 1024
```

## Configuration

### Alert Rule Definition

```yaml
alerts:
  - name: "high-blocked-goroutines"
    expr: "blocked_goroutines > 150"
    for: "5s"
    sinks:
      - "log"
      - "slack:https://hooks.slack.com/services/..."

  - name: "high-heap-usage"
    expr: "heap_bytes / 1024 / 1024 > 512"
    for: "10s"
    sinks:
      - "log"
```

### Available Metrics

1. **Goroutine Metrics**

   - `blocked_goroutines` - Number of blocked goroutines
   - `running_goroutines` - Number of running goroutines
   - `total_goroutines` - Total number of goroutines

2. **Memory Metrics**

   - `heap_bytes` - Total heap bytes
   - `heap_objects` - Number of heap objects
   - `stack_bytes` - Stack memory usage

3. **GC Metrics**
   - `gc_pause_ns` - Last GC pause duration
   - `gc_count` - Total GC count
   - `total_pause_ns` - Total GC pause time

## Best Practices

### Rule Design

1. **Thresholds**

   - Set realistic thresholds
   - Consider system capacity
   - Account for normal variations

2. **Duration**

   - Use appropriate `for` duration
   - Avoid too short durations
   - Consider alert fatigue

3. **Complexity**
   - Keep expressions simple
   - Use parentheses for clarity
   - Document complex rules

### Performance

1. **Evaluation**

   - Rules are evaluated frequently
   - Keep expressions efficient
   - Avoid complex calculations

2. **Resource Usage**
   - Monitor rule evaluation overhead
   - Limit number of active rules
   - Use appropriate sampling rate

## Integration

### Alert Sinks

1. **Log Sink**

   ```yaml
   sinks:
     - "log"
   ```

2. **Slack Integration**

   ```yaml
   sinks:
     - "slack:https://hooks.slack.com/services/..."
   ```

3. **Webhook**

   ```yaml
   sinks:
     - "webhook:https://example.com/webhook"
   ```

4. **Jira Integration**
   ```yaml
   sinks:
     - "jira:https://your-domain.atlassian.net"
   ```

### Custom Sinks

Implement the `Sink` interface:

```go
type Sink interface {
    Notify(ruleName, msg string)
}
```

## Troubleshooting

### Common Issues

1. **Syntax Errors**

   - Check operator precedence
   - Verify parentheses matching
   - Validate metric names

2. **False Positives**

   - Adjust thresholds
   - Increase duration
   - Review conditions

3. **Performance Issues**
   - Simplify expressions
   - Reduce evaluation frequency
   - Monitor system load

## Future Improvements

1. **Language Features**

   - Time-based functions
   - Statistical functions
   - Custom functions

2. **Integration**

   - More sink types
   - Custom metrics
   - External data sources

3. **Tooling**
   - Rule validation
   - Expression testing
   - Performance analysis
