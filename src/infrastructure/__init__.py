# Infrastructure Layer
#
# Adapters for external concerns. Implements ports/ interfaces.
# This is the outermost layer — depends on application/ and domain/.
#
# Contains:
#   persistence/ — Database, file storage implementations
#   messaging/   — Message bus, event publishing implementations
#   external/    — External API clients, third-party integrations
